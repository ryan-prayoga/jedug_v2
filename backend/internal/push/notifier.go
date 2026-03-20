package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

const (
	pushWorkerCount    = 2
	pushClaimBatchSize = 20
	pushLockTimeout    = 2 * time.Minute
	pushPollInterval   = 5 * time.Second
	pushMaxAttempts    = 5
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
	jobRepo    repository.PushDeliveryJobRepository
	httpClient *http.Client
	wake       chan struct{}
}

type notificationPayload struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	IssueID string `json:"issue_id"`
	URL     string `json:"url"`
	Type    string `json:"type"`
}

type sendOutcome struct {
	retryable bool
	err       error
}

func NewNotifier(
	cfg Config,
	repo repository.PushSubscriptionRepository,
	jobRepo repository.PushDeliveryJobRepository,
) *Notifier {
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
		repo:    repo,
		jobRepo: jobRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		wake: make(chan struct{}, 1),
	}
	if notifier.cfg.Enabled {
		notifier.startWorkers(pushWorkerCount)
		notifier.wakeWorkers()
	}
	return notifier
}

func (n *Notifier) DeliverBatch(ctx context.Context, deliveries []repository.PushDelivery) error {
	if !n.cfg.Enabled || len(deliveries) == 0 {
		return nil
	}

	if err := n.jobRepo.EnqueueBatch(ctx, deliveries); err != nil {
		return fmt.Errorf("enqueue push deliveries: %w", err)
	}

	n.wakeWorkers()
	return nil
}

func (n *Notifier) send(ctx context.Context, subscription *domain.PushSubscription, payload []byte, issueID uuid.UUID) sendOutcome {
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
		return sendOutcome{
			retryable: true,
			err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
		return sendOutcome{}
	}

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		if disableErr := n.repo.DisableByEndpoint(ctx, subscription.Endpoint); disableErr != nil {
			log.Printf("[PUSH] disable_subscription_error endpoint=%s error=%v", endpointHost(subscription.Endpoint), disableErr)
		}
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return sendOutcome{
		retryable: isRetryablePushStatus(resp.StatusCode),
		err: fmt.Errorf(
			"web push responded %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		),
	}
}

func issueURL(siteURL string, issueID uuid.UUID) string {
	return siteURL + "/issues/" + issueID.String()
}

func (n *Notifier) startWorkers(workerCount int) {
	for workerIndex := 0; workerIndex < workerCount; workerIndex++ {
		go n.workerLoop()
	}
}

func (n *Notifier) wakeWorkers() {
	select {
	case n.wake <- struct{}{}:
	default:
	}
}

func (n *Notifier) workerLoop() {
	ticker := time.NewTicker(pushPollInterval)
	defer ticker.Stop()

	for {
		n.processPendingJobs()

		select {
		case <-ticker.C:
		case <-n.wake:
		}
	}
}

func (n *Notifier) processPendingJobs() {
	for {
		claimCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		jobs, err := n.jobRepo.ClaimBatch(claimCtx, pushClaimBatchSize, pushLockTimeout)
		cancel()
		if err != nil {
			log.Printf("[PUSH] claim_jobs_error error=%v", err)
			return
		}
		if len(jobs) == 0 {
			return
		}

		n.deliverClaimedJobs(jobs)
		if len(jobs) < pushClaimBatchSize {
			return
		}
	}
}

func (n *Notifier) deliverClaimedJobs(jobs []*repository.PushDeliveryJob) {
	followerIDs := uniqueFollowerIDsFromJobs(jobs)

	loadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	subscriptionsByFollower, err := n.repo.GetActiveByFollowerIDs(loadCtx, followerIDs)
	cancel()
	if err != nil {
		log.Printf("[PUSH] load_subscriptions_error error=%v", err)
		for _, job := range jobs {
			n.retryOrFail(job, fmt.Sprintf("load_subscriptions_error: %v", err))
		}
		return
	}

	for _, job := range jobs {
		subscriptions := subscriptionsByFollower[job.FollowerID]
		if len(subscriptions) == 0 {
			if err := n.jobRepo.MarkFailed(context.Background(), job.ID, "no active push subscriptions"); err != nil {
				log.Printf("[PUSH] mark_failed_error job=%s error=%v", job.ID, err)
			}
			continue
		}

		payload, err := json.Marshal(notificationPayload{
			Title:   job.Title,
			Body:    job.Message,
			IssueID: job.IssueID.String(),
			URL:     issueURL(n.cfg.SiteURL, job.IssueID),
			Type:    job.Type,
		})
		if err != nil {
			log.Printf("[PUSH] payload_marshal_error follower=%s issue=%s error=%v", job.FollowerID, job.IssueID, err)
			if markErr := n.jobRepo.MarkFailed(context.Background(), job.ID, fmt.Sprintf("payload_marshal_error: %v", err)); markErr != nil {
				log.Printf("[PUSH] mark_failed_error job=%s error=%v", job.ID, markErr)
			}
			continue
		}

		successCount := 0
		retryableErrors := make([]string, 0)
		permanentErrors := make([]string, 0)

		for _, subscription := range subscriptions {
			sendCtx, sendCancel := context.WithTimeout(context.Background(), 12*time.Second)
			outcome := n.send(sendCtx, subscription, payload, job.IssueID)
			sendCancel()

			if outcome.err == nil {
				successCount++
				continue
			}

			log.Printf(
				"[PUSH] send_error follower=%s issue=%s endpoint=%s error=%v",
				job.FollowerID,
				job.IssueID,
				endpointHost(subscription.Endpoint),
				outcome.err,
			)

			if outcome.retryable {
				retryableErrors = append(retryableErrors, outcome.err.Error())
			} else {
				permanentErrors = append(permanentErrors, outcome.err.Error())
			}
		}

		if successCount > 0 {
			if err := n.jobRepo.MarkDelivered(context.Background(), job.ID); err != nil {
				log.Printf("[PUSH] mark_delivered_error job=%s error=%v", job.ID, err)
			}
			continue
		}

		lastError := strings.Join(append(retryableErrors, permanentErrors...), " | ")
		if len(retryableErrors) > 0 {
			n.retryOrFail(job, lastError)
			continue
		}
		if lastError == "" {
			lastError = "push delivery failed without provider response"
		}
		if err := n.jobRepo.MarkFailed(context.Background(), job.ID, lastError); err != nil {
			log.Printf("[PUSH] mark_failed_error job=%s error=%v", job.ID, err)
		}
	}
}

func (n *Notifier) retryOrFail(job *repository.PushDeliveryJob, lastError string) {
	if job.AttemptCount >= pushMaxAttempts {
		if err := n.jobRepo.MarkFailed(context.Background(), job.ID, lastError); err != nil {
			log.Printf("[PUSH] mark_failed_error job=%s error=%v", job.ID, err)
		}
		return
	}

	nextAttemptAt := time.Now().Add(nextRetryDelay(job.AttemptCount))
	if err := n.jobRepo.MarkRetry(context.Background(), job.ID, nextAttemptAt, lastError); err != nil {
		log.Printf("[PUSH] mark_retry_error job=%s error=%v", job.ID, err)
	}
}

func nextRetryDelay(attemptCount int) time.Duration {
	switch {
	case attemptCount <= 1:
		return 30 * time.Second
	case attemptCount == 2:
		return 2 * time.Minute
	case attemptCount == 3:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func isRetryablePushStatus(statusCode int) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	return statusCode >= 500
}

func uniqueFollowerIDsFromJobs(jobs []*repository.PushDeliveryJob) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(jobs))
	result := make([]uuid.UUID, 0, len(jobs))
	for _, job := range jobs {
		if _, ok := seen[job.FollowerID]; ok {
			continue
		}
		seen[job.FollowerID] = struct{}{}
		result = append(result, job.FollowerID)
	}
	return result
}

func endpointHost(endpoint string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	if parsed.Host == "" {
		return endpoint
	}
	return parsed.Host
}
