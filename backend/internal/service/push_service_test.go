package service

import "testing"

func TestValidatePushEndpointRejectsUnexpectedHost(t *testing.T) {
	if _, err := validatePushEndpoint("https://example.com/webpush/123"); err == nil {
		t.Fatalf("expected unexpected host to be rejected")
	}
}

func TestValidatePushEndpointRejectsInsecureScheme(t *testing.T) {
	if _, err := validatePushEndpoint("http://fcm.googleapis.com/fcm/send/123"); err == nil {
		t.Fatalf("expected insecure scheme to be rejected")
	}
}

func TestValidatePushEndpointAcceptsFCM(t *testing.T) {
	normalized, err := validatePushEndpoint("https://fcm.googleapis.com/fcm/send/abc123")
	if err != nil {
		t.Fatalf("expected valid FCM endpoint, got %v", err)
	}
	if normalized != "https://fcm.googleapis.com/fcm/send/abc123" {
		t.Fatalf("unexpected normalized endpoint: %s", normalized)
	}
}
