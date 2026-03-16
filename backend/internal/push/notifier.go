package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type Config struct {
	Enabled         bool
	SiteURL         string
	Subscriber      string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	TTLSeconds      int
}

type Notifier struct {
	cfg        Config
	repo       repository.PushSubscriptionRepository
	httpClient *http.Client
	jobs       chan []repository.PushDelivery
}

type notificationPayload struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	IssueID string `json:"issue_id"`
	URL     string `json:"url"`
	Type    string `json:"type"`
}

func NewNotifier(cfg Config, repo repository.PushSubscriptionRepository) *Notifier {
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		ttl = 300
	}

	notifier := &Notifier{
		cfg: Config{
			Enabled:         cfg.Enabled,
			SiteURL:         strings.TrimRight(cfg.SiteURL, "/"),
			Subscriber:      strings.TrimSpace(cfg.Subscriber),
			VAPIDPublicKey:  strings.TrimSpace(cfg.VAPIDPublicKey),
			VAPIDPrivateKey: strings.TrimSpace(cfg.VAPIDPrivateKey),
			TTLSeconds:      ttl,
		},
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		jobs: make(chan []repository.PushDelivery, 128),
	}
	if notifier.cfg.Enabled {
		notifier.startWorkers(2)
	}
	return notifier
}

func (n *Notifier) DeliverBatch(ctx context.Context, deliveries []repository.PushDelivery) error {
	if !n.cfg.Enabled || len(deliveries) == 0 {
		return nil
	}

	job := append(make([]repository.PushDelivery, 0, len(deliveries)), deliveries...)
	select {
	case n.jobs <- job:
		return nil
	default:
		return fmt.Errorf("push delivery queue is full")
	}
}

func (n *Notifier) send(ctx context.Context, subscription *domain.PushSubscription, payload []byte, issueID uuid.UUID) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			Auth:   subscription.Auth,
			P256dh: subscription.P256DH,
		},
	}, &webpush.Options{
		HTTPClient:      n.httpClient,
		Subscriber:      n.cfg.Subscriber,
		TTL:             n.cfg.TTLSeconds,
		Topic:           "issue-" + issueID.String(),
		Urgency:         webpush.UrgencyHigh,
		VAPIDPublicKey:  n.cfg.VAPIDPublicKey,
		VAPIDPrivateKey: n.cfg.VAPIDPrivateKey,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
		return nil
	}

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		if disableErr := n.repo.DisableByEndpoint(ctx, subscription.Endpoint); disableErr != nil {
			log.Printf("[PUSH] disable_subscription_error endpoint=%s error=%v", subscription.Endpoint, disableErr)
		}
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("web push responded %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func issueURL(siteURL string, issueID uuid.UUID) string {
	return siteURL + "/issues/" + issueID.String()
}

func uniqueFollowerIDs(deliveries []repository.PushDelivery) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(deliveries))
	result := make([]uuid.UUID, 0, len(deliveries))
	for _, delivery := range deliveries {
		if _, ok := seen[delivery.FollowerID]; ok {
			continue
		}
		seen[delivery.FollowerID] = struct{}{}
		result = append(result, delivery.FollowerID)
	}
	return result
}

func (n *Notifier) startWorkers(workerCount int) {
	for workerIndex := 0; workerIndex < workerCount; workerIndex++ {
		go n.workerLoop()
	}
}

func (n *Notifier) workerLoop() {
	for deliveries := range n.jobs {
		jobCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		n.deliverBatchSync(jobCtx, deliveries)
		cancel()
	}
}

func (n *Notifier) deliverBatchSync(ctx context.Context, deliveries []repository.PushDelivery) {
	followerIDs := uniqueFollowerIDs(deliveries)
	subscriptionsByFollower, err := n.repo.GetActiveByFollowerIDs(ctx, followerIDs)
	if err != nil {
		log.Printf("[PUSH] load_subscriptions_error error=%v", err)
		return
	}

	for _, delivery := range deliveries {
		subscriptions := subscriptionsByFollower[delivery.FollowerID]
		if len(subscriptions) == 0 {
			continue
		}

		payload, err := json.Marshal(notificationPayload{
			Title:   delivery.Title,
			Body:    delivery.Message,
			IssueID: delivery.IssueID.String(),
			URL:     issueURL(n.cfg.SiteURL, delivery.IssueID),
			Type:    delivery.Type,
		})
		if err != nil {
			log.Printf("[PUSH] payload_marshal_error follower=%s issue=%s error=%v", delivery.FollowerID, delivery.IssueID, err)
			continue
		}

		for _, subscription := range subscriptions {
			if sendErr := n.send(ctx, subscription, payload, delivery.IssueID); sendErr != nil {
				log.Printf("[PUSH] send_error follower=%s issue=%s endpoint=%s error=%v",
					delivery.FollowerID,
					delivery.IssueID,
					subscription.Endpoint,
					sendErr,
				)
			}
		}
	}
}
