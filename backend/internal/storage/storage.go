package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxFileSizeBytes = 20 * 1024 * 1024 // 20 MB

var (
	allowedContentTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/heic": ".heic",
		"image/heif": ".heif",
	}
	allowedExtensions = map[string]struct{}{
		".jpg":  {},
		".jpeg": {},
		".png":  {},
		".webp": {},
		".heic": {},
		".heif": {},
	}
)

type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

func newValidationError(message string) error {
	return &ValidationError{message: message}
}

func IsValidationError(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}

type UploadRequest struct {
	Filename    string
	ContentType string
	SizeBytes   int
}

type PresignResult struct {
	ObjectKey    string            `json:"object_key"`
	UploadMode   string            `json:"upload_mode"`
	UploadURL    string            `json:"upload_url"`
	UploadMethod string            `json:"upload_method,omitempty"`
	PublicURL    string            `json:"public_url"`
	Headers      map[string]string `json:"headers,omitempty"`
}

type ObjectInfo struct {
	SizeBytes   int64
	ContentType string
}

type Driver interface {
	Name() string
	GenerateObjectKey(req UploadRequest) (string, error)
	CreatePresign(ctx context.Context, req UploadRequest, objectKey string) (*PresignResult, error)
	BuildPublicURL(objectKey string) string
	Upload(ctx context.Context, objectKey, contentType string, body []byte) error
	Stat(ctx context.Context, objectKey string) (*ObjectInfo, error)
}

func AllowedContentTypes() []string {
	values := make([]string, 0, len(allowedContentTypes))
	for contentType := range allowedContentTypes {
		values = append(values, contentType)
	}
	slices.Sort(values)
	return values
}

func NormalizeContentType(contentType string) string {
	base := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.Index(base, ";"); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}
	return base
}

func ValidateContentType(contentType string) error {
	if _, ok := allowedContentTypes[NormalizeContentType(contentType)]; !ok {
		return newValidationError(
			"unsupported content type; allowed: " + strings.Join(AllowedContentTypes(), ", "),
		)
	}
	return nil
}

func ValidateSizeBytes(sizeBytes int) error {
	if sizeBytes <= 0 || sizeBytes > MaxFileSizeBytes {
		return newValidationError(
			fmt.Sprintf("size_bytes must be between 1 and %d", MaxFileSizeBytes),
		)
	}
	return nil
}

func ValidateUploadRequest(req UploadRequest) error {
	if strings.TrimSpace(req.Filename) == "" {
		return newValidationError("filename is required")
	}
	if err := ValidateContentType(req.ContentType); err != nil {
		return err
	}
	if err := ValidateSizeBytes(req.SizeBytes); err != nil {
		return err
	}
	return nil
}

func ValidateSubmittedMedia(objectKey, contentType string, sizeBytes int) error {
	if err := ValidateObjectKey(objectKey); err != nil {
		return err
	}
	if err := ValidateContentType(contentType); err != nil {
		return err
	}
	if err := ValidateSizeBytes(sizeBytes); err != nil {
		return err
	}
	if ext, ok := allowedContentTypes[NormalizeContentType(contentType)]; ok && !hasMatchingExtension(objectKey, ext) {
		return newValidationError("object_key extension does not match mime_type")
	}
	return nil
}

func ValidateObjectKey(objectKey string) error {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return newValidationError("object_key is required")
	}
	if isAbsoluteURL(key) {
		return newValidationError("object_key must be a storage key, not a URL")
	}
	if strings.Contains(key, "..") {
		return newValidationError("invalid object key")
	}

	normalized := NormalizeObjectKey(key)
	if normalized == "" || normalized == "." || strings.HasPrefix(normalized, "/") {
		return newValidationError("invalid object key")
	}
	if strings.Contains(normalized, " ") {
		return newValidationError("object_key must not contain spaces")
	}
	if normalized != strings.ToLower(normalized) {
		return newValidationError("object_key must be lowercase")
	}
	if path.Clean(normalized) != normalized {
		return newValidationError("invalid object key")
	}
	if !strings.HasPrefix(normalized, "issues/") {
		return newValidationError("object_key must start with issues/")
	}
	parts := strings.Split(normalized, "/")
	if len(parts) != 4 {
		return newValidationError("object_key must follow issues/YYYY/MM/name.ext")
	}
	if len(parts[1]) != 4 || strings.Trim(parts[1], "0123456789") != "" {
		return newValidationError("object_key must use a 4-digit year directory")
	}
	if len(parts[2]) != 2 || strings.Trim(parts[2], "0123456789") != "" {
		return newValidationError("object_key must use a 2-digit month directory")
	}
	ext := strings.ToLower(path.Ext(normalized))
	if _, ok := allowedExtensions[ext]; !ok {
		return newValidationError("object_key has unsupported extension")
	}
	return nil
}

func NormalizeObjectKey(objectKey string) string {
	key := strings.TrimSpace(strings.ReplaceAll(objectKey, "\\", "/"))
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimPrefix(key, "uploads/gallery/")
	key = strings.TrimPrefix(key, "/uploads/gallery/")
	key = path.Clean("/" + key)
	return strings.TrimPrefix(key, "/")
}

func NewObjectKey(contentType string, now time.Time) (string, error) {
	ext, ok := allowedContentTypes[NormalizeContentType(contentType)]
	if !ok {
		return "", newValidationError(
			"unsupported content type; allowed: " + strings.Join(AllowedContentTypes(), ", "),
		)
	}

	timestamp := now.UTC()
	return fmt.Sprintf(
		"issues/%04d/%02d/%s%s",
		timestamp.Year(),
		int(timestamp.Month()),
		uuid.NewString(),
		ext,
	), nil
}

func hasMatchingExtension(objectKey, expected string) bool {
	current := strings.ToLower(path.Ext(NormalizeObjectKey(objectKey)))
	if current == expected {
		return true
	}
	return expected == ".jpg" && current == ".jpeg"
}

func isAbsoluteURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

type LocalDriver struct {
	publicBaseURL string
	baseDir       string
}

func NewLocalDriver(publicBaseURL, baseDir string) *LocalDriver {
	return &LocalDriver{
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		baseDir:       baseDir,
	}
}

func (d *LocalDriver) Name() string {
	return "local"
}

func (d *LocalDriver) GenerateObjectKey(req UploadRequest) (string, error) {
	return NewObjectKey(req.ContentType, time.Now().UTC())
}

func (d *LocalDriver) CreatePresign(_ context.Context, req UploadRequest, objectKey string) (*PresignResult, error) {
	return &PresignResult{
		ObjectKey:    objectKey,
		UploadMode:   d.Name(),
		UploadURL:    "/api/v1/uploads/file/" + objectKey,
		UploadMethod: "POST",
		PublicURL:    d.BuildPublicURL(objectKey),
		Headers: map[string]string{
			"Content-Type": NormalizeContentType(req.ContentType),
		},
	}, nil
}

func (d *LocalDriver) BuildPublicURL(objectKey string) string {
	key := NormalizeObjectKey(objectKey)
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s/uploads/gallery/%s", d.publicBaseURL, key)
}

func (d *LocalDriver) Upload(_ context.Context, objectKey, _ string, body []byte) error {
	absPath := d.AbsPath(objectKey)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absPath, body, 0o644)
}

func (d *LocalDriver) Stat(_ context.Context, objectKey string) (*ObjectInfo, error) {
	info, err := os.Stat(d.AbsPath(objectKey))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return &ObjectInfo{
		SizeBytes: info.Size(),
	}, nil
}

func (d *LocalDriver) AbsPath(objectKey string) string {
	return filepath.Join(d.baseDir, filepath.FromSlash(NormalizeObjectKey(objectKey)))
}

func (d *LocalDriver) Exists(objectKey string) bool {
	_, err := os.Stat(d.AbsPath(objectKey))
	return err == nil
}
