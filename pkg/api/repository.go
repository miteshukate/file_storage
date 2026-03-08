package api

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
)

type Filter interface{}

type ContentRepository interface {
	FindAll() ([]Content, error)
	FindById(id primitive.ObjectID) (*Content, error)
	FindByParentId(parentId string) (*Content, error)
	FindByFilter(filter interface{}) ([]Content, error)
	Create(content *Content) (*Content, error)
	Update(id primitive.ObjectID, filter Filter, content *Content) (*Content, error)
	FindByIds(ids []string) ([]Content, error)
}

// SearchResult represents a search result
type SearchResult struct {
	ID       string  `json:"id"`
	Score    float64 `json:"_score"`
	Filename string  `json:"filename"`
}

// SearchService defines the interface for full-text search operations
type SearchService interface {
	// EnsureIndex creates the search index if it doesn't exist
	EnsureIndex(ctx context.Context) error
	// IndexDocument indexes a document for searching
	IndexDocument(ctx context.Context, doc *OpenSearchIndexDocument) error
	// IndexDocumentWithExtraction extracts text from a reader and indexes the document
	IndexDocumentWithExtraction(ctx context.Context, doc *OpenSearchIndexDocument, reader io.ReadCloser) error
	// SearchByKeyword searches for documents matching a keyword
	SearchByKeyword(ctx context.Context, keyword string, limit int) ([]SearchResult, error)
	// DeleteFromIndex removes a document from the search index
	DeleteFromIndex(ctx context.Context, docID string) error
	// FlushIndex forces an index flush
	FlushIndex(ctx context.Context) error
}
