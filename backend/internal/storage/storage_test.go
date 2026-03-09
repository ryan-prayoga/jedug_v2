package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeDriver struct {
	name string
}

func (d fakeDriver) Name() string {
	return d.name
}

func (d fakeDriver) GenerateObjectKey(_ UploadRequest) (string, error) {
	return "issues/2026/03/fake.webp", nil
}

func (d fakeDriver) CreatePresign(_ context.Context, req UploadRequest, objectKey string) (*PresignResult, error) {
	return &PresignResult{
		ObjectKey:    objectKey,
		UploadMode:   d.name,
		UploadURL:    "https://upload.example.test/" + objectKey,
		UploadMethod: "PUT",
		PublicURL:    d.BuildPublicURL(objectKey),
		Headers: map[string]string{
			"Content-Type": NormalizeContentType(req.ContentType),
		},
	}, nil
}

func (d fakeDriver) BuildPublicURL(objectKey string) string {
	return "https://cdn.example.test/" + NormalizeObjectKey(objectKey)
}

func TestValidateObjectKey(t *testing.T) {
	t.Parallel()

	valid := "issues/2026/03/7dbb7450-c365-4988-ad5f-534fb24b9bec.webp"
	if err := ValidateObjectKey(valid); err != nil {
		t.Fatalf("expected valid key, got %v", err)
	}

	invalid := []string{
		"",
		"https://cdn.example.test/issues/2026/03/file.webp",
		"issues/2026/03/File.webp",
		"issues/2026/03/file with space.webp",
		"../issues/2026/03/file.webp",
		"issues/2026/03/file.exe",
	}
	for _, candidate := range invalid {
		if err := ValidateObjectKey(candidate); err == nil {
			t.Fatalf("expected %q to be invalid", candidate)
		}
	}
}

func TestServiceResolvePublicURLFallsBackToLegacyLocal(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	local := NewLocalDriver("https://api.example.test", tmpDir)

	objectKey := "issues/2026/03/legacy.webp"
	absPath := filepath.Join(tmpDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(absPath, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	service := NewService(fakeDriver{name: "r2"}, local)

	got := service.ResolvePublicURL(objectKey)
	want := "https://api.example.test/uploads/gallery/issues/2026/03/legacy.webp"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	got = service.ResolvePublicURL("issues/2026/03/new.webp")
	want = "https://cdn.example.test/issues/2026/03/new.webp"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalDriverCreatePresign(t *testing.T) {
	t.Parallel()

	driver := NewLocalDriver("https://api.example.test", "/tmp/uploads")
	result, err := driver.CreatePresign(context.Background(), UploadRequest{
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   1024,
	}, "issues/2026/03/file.webp")
	if err != nil {
		t.Fatalf("CreatePresign returned error: %v", err)
	}
	if result.UploadMethod != "POST" {
		t.Fatalf("expected POST, got %q", result.UploadMethod)
	}
	if result.UploadURL != "/api/v1/uploads/file/issues/2026/03/file.webp" {
		t.Fatalf("unexpected upload URL: %q", result.UploadURL)
	}
}

func TestR2DriverCreatePresign(t *testing.T) {
	t.Parallel()

	driver, err := NewR2Driver(context.Background(), R2Config{
		AccountID:       "acct",
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
		Bucket:          "jedug-media",
		Endpoint:        "https://acct.r2.cloudflarestorage.com",
		PublicBaseURL:   "https://media.example.test",
	})
	if err != nil {
		t.Fatalf("NewR2Driver returned error: %v", err)
	}

	result, err := driver.CreatePresign(context.Background(), UploadRequest{
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   2048,
	}, "issues/2026/03/file.webp")
	if err != nil {
		t.Fatalf("CreatePresign returned error: %v", err)
	}
	if result.UploadMethod != "PUT" {
		t.Fatalf("expected PUT, got %q", result.UploadMethod)
	}
	if !strings.Contains(result.UploadURL, "/jedug-media/issues/2026/03/file.webp") {
		t.Fatalf("unexpected presigned URL: %q", result.UploadURL)
	}
	if result.PublicURL != "https://media.example.test/issues/2026/03/file.webp" {
		t.Fatalf("unexpected public URL: %q", result.PublicURL)
	}
}
