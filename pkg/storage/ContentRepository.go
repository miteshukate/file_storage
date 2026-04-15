package storage

// ContentRepository defines all persistence operations for Content metadata.
type ContentRepository interface {
	FindAll() ([]Content, error)
	FindById(id int64) (*Content, error)
	FindByParentId(parentId int64) ([]Content, error)
	FindByIds(ids []int64) ([]Content, error)
	FindByFilter(filter map[string]interface{}) ([]Content, error)
	Create(content *Content) (*Content, error)
	Update(id int64, content *Content) (*Content, error)
	Delete(id int64) error
}
