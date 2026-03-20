package sse

import "testing"

func TestFormatEventIncludesOptionalID(t *testing.T) {
	msg := FormatEvent("notification", []byte(`{"ok":true}`), "123")
	if msg != "id: 123\nevent: notification\ndata: {\"ok\":true}\n\n" {
		t.Fatalf("unexpected SSE frame: %q", msg)
	}
}
