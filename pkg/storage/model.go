package storage

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

// bun model for Users
type User struct {
	UserId       int64  `bun:"user_id,pk,autoincrement" json:"userId"`
	Email        string `bun:"email,unique" json:"email"`
	PasswordHash string `bun:"password_hash" json:"-"`
	CreatedAt    int64  `bun:"created_at" json:"createdAt"`
	UpdatedAt    int64  `bun:"updated_at" json:"updatedAt"`
}

// Content represents object metadata stored in PostgreSQL via bun.
type Content struct {
	bun.BaseModel `bun:"table:contents,alias:c"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	ParentID     *int64    `bun:"parent_id,nullzero"          json:"parentId,omitempty"`
	Name         string    `bun:"name,notnull"                json:"name"`
	Type         string    `bun:"type,notnull,default:'file'" json:"type,omitempty"`
	Size         int64     `bun:"size,default:0"              json:"size,omitempty"`
	ContentType  string    `bun:"content_type"                json:"contentType,omitempty"`
	Status       string    `bun:"status,notnull,default:'Active'" json:"status,omitempty"`
	ContentId    uuid.UUID `bun:"content_id,pk,autoincrement" json:"contentId"`
	ETag         string    `bun:"etag"                        json:"etag,omitempty"`
	LastModified time.Time `bun:"last_modified"               json:"lastModified,omitempty"`
	CreatedAt    time.Time `bun:"created_at,notnull"          json:"createdAt,omitempty"`
	UpdatedAt    time.Time `bun:"updated_at,notnull"          json:"updatedAt,omitempty"`
}

// OpenSearchIndexDocument represents a searchable document in the OpenSearch index
type OpenSearchIndexDocument struct {
	ID            string    `json:"id"` // MongoDB ObjectID
	Filename      string    `json:"filename"`
	ContentType   string    `json:"contentType"`
	ExtractedText string    `json:"extractedText"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	FileSize      int64     `json:"fileSize"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID       string  `json:"id"`
	Score    float64 `json:"_score"`
	Filename string  `json:"filename"`
}
