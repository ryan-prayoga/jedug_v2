package storage

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Driver abstracts file storage.
// Current implementation: LocalDriver (filesystem).
// Future: swap in R2Driver without changing any caller.
type Driver interface {
	GenerateObjectKey(filename string) string
	BuildPublicURL(objectKey string) string
	BuildUploadURL(objectKey string) string
	UploadMode() string
}

// LocalDriver stores files on the local filesystem.
// This is a temporary implementation until Cloudflare R2 credentials are available.
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

func (d *LocalDriver) GenerateObjectKey(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	now := time.Now().UTC()
	return fmt.Sprintf("issues/%d/%02d/%s%s", now.Year(), int(now.Month()), uuid.New().String(), ext)
}

func (d *LocalDriver) BuildPublicURL(objectKey string) string {
	return fmt.Sprintf("%s/uploads/gallery/%s", d.publicBaseURL, objectKey)
}

// BuildUploadURL returns the endpoint the client should POST the raw file body to.
func (d *LocalDriver) BuildUploadURL(objectKey string) string {
	return "/api/v1/uploads/file/" + objectKey
}

func (d *LocalDriver) UploadMode() string { return "local" }

// AbsPath resolves the absolute filesystem path for a given object key.
func (d *LocalDriver) AbsPath(objectKey string) string {
	return filepath.Join(d.baseDir, filepath.FromSlash(objectKey))
}
