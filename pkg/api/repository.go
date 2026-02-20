package api

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Filter interface{}

type ContentRepository interface {
	FindAll() ([]Content, error)
	FindById(id uuid.UUID) (*Content, error)
	FindByParentId(parentId string) (*Content, error)
	FindByFilter(filter interface{}) ([]Content, error)
	Create(content *Content) (*Content, error)
	Update(id primitive.ObjectID, filter Filter, content *Content) (*Content, error)
}
