package api

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Content represents object metadata used across storage and repositories.
type Content struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ParentID     string             `bson:"parentId,omitempty" json:"parentId,omitempty"`
	Name         string             `bson:"name" json:"name"`
	Type         string             `bson:"type,omitempty" json:"type,omitempty"` // e.g., file, folder
	Size         int64              `bson:"size,omitempty" json:"size,omitempty"`
	ContentType  string             `bson:"contentType,omitempty" json:"contentType,omitempty"`
	Status       string             `bson:"status,omitempty" json:"status,omitempty"`
	ETag         string             `bson:"etag,omitempty" json:"etag,omitempty"`
	LastModified time.Time          `bson:"lastModified,omitempty" json:"lastModified,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt    time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
