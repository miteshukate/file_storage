package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"file_storage/pkg/api"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewMinioClient creates a MinIO client using provided connection parameters.
func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	}
	return minio.New(endpoint, opts)
}

// MinioDocumentService implements fetching documents from MinIO.
type MinioDocumentService struct {
	client *minio.Client
	bucket string
	prefix string
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getbool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

// NewMinioDocumentService constructs a MinioDocumentService from an existing client.
func NewMinioDocumentService() *MinioDocumentService {

	bucket := getenv("MINIO_BUCKET", "test")
	prefix := getenv("MINIO_PREFIX", "")
	endpoint := getenv("MINIO_ENDPOINT", "127.0.0.1:9000")
	accessKey := getenv("MINIO_ACCESS_KEY", "admin")
	secretKey := getenv("MINIO_SECRET_KEY", "ifyouusethispasswordsupportwilllaughatyou")
	useSSL := getbool("MINIO_USE_SSL", false)

	client, err := NewMinioClient(endpoint, accessKey, secretKey, useSSL)
	if err != nil {
		log.Fatalf("failed to create minio client: %v", err)
	}
	return &MinioDocumentService{client: client, bucket: bucket, prefix: prefix}
}

// StreamDocument returns a reader for the object and its metadata.
func (s *MinioDocumentService) StreamDocument(ctx context.Context, name string) (io.ReadCloser, api.Content, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, s.effectiveKey(name), minio.GetObjectOptions{})
	if err != nil {
		return nil, api.Content{}, err
	}

	st, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, api.Content{}, err
	}
	md := api.Content{
		Name:         st.Key,
		Size:         st.Size,
		ETag:         st.ETag,
		LastModified: st.LastModified,
		ContentType:  st.ContentType,
	}
	return obj, md, nil
}

// GetPresignedURL returns a presigned URL.
func (s *MinioDocumentService) GetPresignedURL(ctx context.Context, name string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, s.effectiveKey(name), expiry, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioDocumentService) effectivePrefix(extra string) string {
	if s.prefix == "" {
		return extra
	}
	if extra == "" {
		return s.prefix
	}
	return fmt.Sprintf("%s/%s", s.prefix, extra)
}

func (s *MinioDocumentService) effectiveKey(name string) string {
	if s.prefix == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", s.prefix, name)
}

func (s *MinioDocumentService) ListDocuments(ctx context.Context) ([]api.Content, error) {
	return make([]api.Content, 0), nil
}

// UploadDocument uploads content to MinIO and returns the server-side ETag and last-modified time if available.
func (s *MinioDocumentService) UploadDocument(ctx context.Context, name string, contentType string, r io.Reader, size int64) (string, time.Time, error) {
	s.client.StatObject(ctx, s.bucket, s.effectiveKey(name), minio.StatObjectOptions{})
	opts := minio.PutObjectOptions{ContentType: contentType}
	info, err := s.client.PutObject(ctx, s.bucket, s.effectiveKey(name), r, size, opts)
	if err != nil {
		return "", time.Time{}, err
	}
	// Head to fetch metadata including last-modified
	st, err := s.client.StatObject(ctx, s.bucket, s.effectiveKey(name), minio.StatObjectOptions{})
	if err != nil {
		// fallback: return zero time if stat fails
		return info.ETag, time.Time{}, nil
	}
	return info.ETag, st.LastModified, nil
}
