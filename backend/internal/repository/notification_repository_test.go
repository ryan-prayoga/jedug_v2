package repository

import "testing"

func TestNotificationEventPreferenceExpr(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{eventType: "photo_added", want: "COALESCE(p.notify_on_photo_added, TRUE)"},
		{eventType: "status_updated", want: "COALESCE(p.notify_on_status_updated, TRUE)"},
		{eventType: "severity_changed", want: "COALESCE(p.notify_on_severity_changed, TRUE)"},
		{eventType: "casualty_reported", want: "COALESCE(p.notify_on_casualty_reported, TRUE)"},
		{eventType: "nearby_issue_created", want: "COALESCE(p.notify_on_nearby_issue_created, TRUE)"},
		{eventType: "issue_created", want: "TRUE"},
	}

	for _, tc := range tests {
		t.Run(tc.eventType, func(t *testing.T) {
			got := notificationEventPreferenceExpr(tc.eventType)
			if got != tc.want {
				t.Fatalf("preference expr mismatch: got %q want %q", got, tc.want)
			}
		})
	}
}
