package push

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type pushSubscriptionRepoFake struct {
	subscriptions map[uuid.UUID][]*domain.PushSubscription
}

func (f *pushSubscriptionRepoFake) Upsert(context.Context, repository.PushSubscriptionUpsertInput) (*domain.PushSubscription, error) {
	return nil, nil
}

func (f *pushSubscriptionRepoFake) Disable(context.Context, uuid.UUID, string) (bool, error) {
	return false, nil
}

func (f *pushSubscriptionRepoFake) DisableByEndpoint(context.Context, string) error {
	return nil
}

func (f *pushSubscriptionRepoFake) CountActiveByFollowerID(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (f *pushSubscriptionRepoFake) GetActiveByFollowerIDs(_ context.Context, followerIDs []uuid.UUID) (map[uuid.UUID][]*domain.PushSubscription, error) {
	result := make(map[uuid.UUID][]*domain.PushSubscription, len(followerIDs))
	for _, followerID := range followerIDs {
		result[followerID] = f.subscriptions[followerID]
	}
	return result, nil
}

type pushDeliveryJobRepoFake struct {
	enqueued      []repository.PushDelivery
	markRetryID   uuid.UUID
	markRetryAt   time.Time
	markRetryErr  string
	markFailedID  uuid.UUID
	markFailedErr string
}

func (f *pushDeliveryJobRepoFake) EnqueueBatch(_ context.Context, deliveries []repository.PushDelivery) error {
	f.enqueued = append(f.enqueued, deliveries...)
	return nil
}

func (f *pushDeliveryJobRepoFake) ClaimBatch(context.Context, int, time.Duration) ([]*repository.PushDeliveryJob, error) {
	return nil, nil
}

func (f *pushDeliveryJobRepoFake) MarkDelivered(context.Context, uuid.UUID) error {
	return nil
}

func (f *pushDeliveryJobRepoFake) MarkRetry(_ context.Context, jobID uuid.UUID, nextAttemptAt time.Time, lastError string) error {
	f.markRetryID = jobID
	f.markRetryAt = nextAttemptAt
	f.markRetryErr = lastError
	return nil
}

func (f *pushDeliveryJobRepoFake) MarkFailed(_ context.Context, jobID uuid.UUID, lastError string) error {
	f.markFailedID = jobID
	f.markFailedErr = lastError
	return nil
}

func TestDeliverBatchEnqueuesOutbox(t *testing.T) {
	jobRepo := &pushDeliveryJobRepoFake{}
	notifier := &Notifier{
		cfg:     Config{Enabled: true},
		repo:    &pushSubscriptionRepoFake{},
		jobRepo: jobRepo,
		wake:    make(chan struct{}, 1),
	}

	delivery := repository.PushDelivery{
		FollowerID: uuid.New(),
		IssueID:    uuid.New(),
		EventID:    42,
		Type:       "status_updated",
		Title:      "Status berubah",
		Message:    "Ada perubahan",
	}
	if err := notifier.DeliverBatch(context.Background(), []repository.PushDelivery{delivery}); err != nil {
		t.Fatalf("DeliverBatch returned error: %v", err)
	}
	if len(jobRepo.enqueued) != 1 {
		t.Fatalf("expected 1 enqueued delivery, got %d", len(jobRepo.enqueued))
	}
	if jobRepo.enqueued[0].EventID != 42 {
		t.Fatalf("expected enqueued event id 42, got %d", jobRepo.enqueued[0].EventID)
	}
}

func TestRetryOrFailSchedulesRetryBelowMaxAttempts(t *testing.T) {
	jobRepo := &pushDeliveryJobRepoFake{}
	notifier := &Notifier{jobRepo: jobRepo}
	job := &repository.PushDeliveryJob{
		ID:           uuid.New(),
		AttemptCount: 2,
	}

	before := time.Now()
	notifier.retryOrFail(job, "provider timeout")

	if jobRepo.markRetryID != job.ID {
		t.Fatalf("expected retry mark for job %s, got %s", job.ID, jobRepo.markRetryID)
	}
	if !jobRepo.markRetryAt.After(before) {
		t.Fatalf("expected next retry in the future, got %v", jobRepo.markRetryAt)
	}
	if jobRepo.markFailedID != uuid.Nil {
		t.Fatalf("did not expect failed mark, got %s", jobRepo.markFailedID)
	}
}

func TestRetryOrFailFailsAtMaxAttempts(t *testing.T) {
	jobRepo := &pushDeliveryJobRepoFake{}
	notifier := &Notifier{jobRepo: jobRepo}
	job := &repository.PushDeliveryJob{
		ID:           uuid.New(),
		AttemptCount: pushMaxAttempts,
	}

	notifier.retryOrFail(job, "provider timeout")

	if jobRepo.markFailedID != job.ID {
		t.Fatalf("expected failed mark for job %s, got %s", job.ID, jobRepo.markFailedID)
	}
	if jobRepo.markRetryID != uuid.Nil {
		t.Fatalf("did not expect retry mark, got %s", jobRepo.markRetryID)
	}
}

func TestDeliverClaimedJobsFailsWithoutActiveSubscriptions(t *testing.T) {
	jobRepo := &pushDeliveryJobRepoFake{}
	followerID := uuid.New()
	job := &repository.PushDeliveryJob{
		ID:         uuid.New(),
		FollowerID: followerID,
		IssueID:    uuid.New(),
		EventID:    99,
		Type:       "status_updated",
		Title:      "Status berubah",
		Message:    "Ada perubahan",
	}
	notifier := &Notifier{
		cfg:     Config{Enabled: true},
		repo:    &pushSubscriptionRepoFake{subscriptions: map[uuid.UUID][]*domain.PushSubscription{}},
		jobRepo: jobRepo,
	}

	notifier.deliverClaimedJobs([]*repository.PushDeliveryJob{job})

	if jobRepo.markFailedID != job.ID {
		t.Fatalf("expected job to be marked failed, got %s", jobRepo.markFailedID)
	}
	if jobRepo.markFailedErr != "no active push subscriptions" {
		t.Fatalf("unexpected failed error: %q", jobRepo.markFailedErr)
	}
}

func TestIsRetryablePushStatus(t *testing.T) {
	if !isRetryablePushStatus(429) {
		t.Fatal("expected 429 to be retryable")
	}
	if !isRetryablePushStatus(503) {
		t.Fatal("expected 503 to be retryable")
	}
	if isRetryablePushStatus(410) {
		t.Fatal("expected 410 to be non-retryable")
	}
}
