package api

import (
	"context"
	"file_storage/pkg/storage"
	"github.com/google/uuid"
	"io"
)

type ContentRepository interface {
	FindAll() ([]storage.Content, error)
	FindById(id int64) (*storage.Content, error)
	FindByContentId(contentId uuid.UUID) (*storage.Content, error)
	FindByParentId(parentId int64) ([]storage.Content, error)
	FindByFilter(filter map[string]interface{}) ([]storage.Content, error)
	Create(content *storage.Content) (*storage.Content, error)
	Update(id int64, content *storage.Content) (*storage.Content, error)
	FindByIds(ids []int64) ([]storage.Content, error)
}

// SearchService defines the interface for full-text search operations
type SearchService interface {
	// EnsureIndex creates the search index if it doesn't exist
	EnsureIndex(ctx context.Context) error
	// IndexDocument indexes a document for searching
	IndexDocument(ctx context.Context, doc *storage.OpenSearchIndexDocument) error
	// IndexDocumentWithExtraction extracts text from a reader and indexes the document
	IndexDocumentWithExtraction(ctx context.Context, doc *storage.OpenSearchIndexDocument, reader io.ReadCloser) error
	// SearchByKeyword searches for documents matching a keyword
	SearchByKeyword(ctx context.Context, keyword string, limit int) ([]storage.SearchResult, error)
	// DeleteFromIndex removes a document from the search index
	DeleteFromIndex(ctx context.Context, docID string) error
	// FlushIndex forces an index flush
	FlushIndex(ctx context.Context) error
}
