package api

import (
	"context"
	"io"
	"time"
)

// StorageService abstracts storage operations so implementations can vary (MinIO, S3, etc.).
type StorageService interface {
	// StreamDocument returns a ReadCloser to stream object content; caller must close.
	StreamDocument(ctx context.Context, name string) (io.ReadCloser, Content, error)
	// GetPresignedURL returns a presigned URL for direct access.
	GetPresignedURL(ctx context.Context, name string, expiry time.Duration) (string, error)
	// UploadDocument uploads content stream to storage under the given name and returns ETag and last-modified.
	UploadDocument(ctx context.Context, name string, contentType string, r io.Reader, size int64) (etag string, lastModified time.Time, err error)
}
