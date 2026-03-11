package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPickBestDuplicateCandidate(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		candidates []duplicateCandidate
		want       duplicateCandidate
	}{
		{
			name: "picks nearest candidate first",
			candidates: []duplicateCandidate{
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    3,
					DistanceM:          22.1,
				},
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    1,
					DistanceM:          7.2,
				},
			},
			want: duplicateCandidate{
				IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Status:             "open",
				VerificationStatus: "pending",
				LastSeenAt:         now,
				SeverityCurrent:    1,
				DistanceM:          7.2,
			},
		},
		{
			name: "uses status priority when distance is equal",
			candidates: []duplicateCandidate{
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					Status:             "in_progress",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    2,
					DistanceM:          10,
				},
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    2,
					DistanceM:          10,
				},
			},
			want: duplicateCandidate{
				IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000004"),
				Status:             "open",
				VerificationStatus: "pending",
				LastSeenAt:         now,
				SeverityCurrent:    2,
				DistanceM:          10,
			},
		},
		{
			name: "uses latest last_seen_at for same distance and status",
			candidates: []duplicateCandidate{
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now.Add(-10 * time.Minute),
					SeverityCurrent:    4,
					DistanceM:          9.5,
				},
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000006"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    1,
					DistanceM:          9.5,
				},
			},
			want: duplicateCandidate{
				IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000006"),
				Status:             "open",
				VerificationStatus: "pending",
				LastSeenAt:         now,
				SeverityCurrent:    1,
				DistanceM:          9.5,
			},
		},
		{
			name: "uses higher severity when previous keys tie",
			candidates: []duplicateCandidate{
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000007"),
					Status:             "open",
					VerificationStatus: "verified",
					LastSeenAt:         now,
					SeverityCurrent:    2,
					DistanceM:          11,
				},
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000008"),
					Status:             "open",
					VerificationStatus: "verified",
					LastSeenAt:         now,
					SeverityCurrent:    4,
					DistanceM:          11,
				},
			},
			want: duplicateCandidate{
				IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000008"),
				Status:             "open",
				VerificationStatus: "verified",
				LastSeenAt:         now,
				SeverityCurrent:    4,
				DistanceM:          11,
			},
		},
		{
			name: "prefers active status over fixed in defensive tie",
			candidates: []duplicateCandidate{
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000009"),
					Status:             "fixed",
					VerificationStatus: "verified",
					LastSeenAt:         now,
					SeverityCurrent:    5,
					DistanceM:          6,
				},
				{
					IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000010"),
					Status:             "open",
					VerificationStatus: "pending",
					LastSeenAt:         now,
					SeverityCurrent:    1,
					DistanceM:          6,
				},
			},
			want: duplicateCandidate{
				IssueID:            uuid.MustParse("00000000-0000-0000-0000-000000000010"),
				Status:             "open",
				VerificationStatus: "pending",
				LastSeenAt:         now,
				SeverityCurrent:    1,
				DistanceM:          6,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pickBestDuplicateCandidate(tc.candidates)
			if got.IssueID != tc.want.IssueID {
				t.Fatalf("issue mismatch: got %s want %s", got.IssueID, tc.want.IssueID)
			}
		})
	}
}

func TestIncomingCasualtyCount(t *testing.T) {
	tests := []struct {
		name  string
		input SubmitInput
		want  int
	}{
		{
			name: "returns zero when has_casualty is false",
			input: SubmitInput{
				HasCasualty:   false,
				CasualtyCount: 5,
			},
			want: 0,
		},
		{
			name: "returns input casualty count when has_casualty is true",
			input: SubmitInput{
				HasCasualty:   true,
				CasualtyCount: 3,
			},
			want: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := incomingCasualtyCount(tc.input)
			if got != tc.want {
				t.Fatalf("casualty count mismatch: got %d want %d", got, tc.want)
			}
		})
	}
}
