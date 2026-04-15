package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// ContentRepositoryImpl implements ContentRepository against PostgreSQL using bun.
type ContentRepositoryImpl struct {
	db *bun.DB
}

func NewContentRepositoryImpl(db *bun.DB) *ContentRepositoryImpl {
	return &ContentRepositoryImpl{db: db}
}

func (r *ContentRepositoryImpl) Create(content *Content) (*Content, error) {
	now := time.Now()
	if content.CreatedAt.IsZero() {
		content.CreatedAt = now
	}
	content.UpdatedAt = now

	_, err := r.db.NewInsert().Model(content).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.Create: %w", err)
	}
	return content, nil
}

func (r *ContentRepositoryImpl) Update(id int64, content *Content) (*Content, error) {
	content.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().
		Model(content).
		Where("content_id = ?", id).
		Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.Update: %w", err)
	}
	return content, nil
}

func (r *ContentRepositoryImpl) FindAll() ([]Content, error) {
	var contents []Content
	err := r.db.NewSelect().Model(&contents).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.FindAll: %w", err)
	}
	return contents, nil
}

func (r *ContentRepositoryImpl) FindById(id int64) (*Content, error) {
	content := &Content{}
	err := r.db.NewSelect().Model(content).Where("content_id = ?", id).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.FindById: %w", err)
	}
	return content, nil
}

func (r *ContentRepositoryImpl) FindByParentId(parentId int64) ([]Content, error) {
	var contents []Content
	err := r.db.NewSelect().Model(&contents).Where("parent_id = ?", parentId).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.FindByParentId: %w", err)
	}
	return contents, nil
}

func (r *ContentRepositoryImpl) FindByIds(ids []int64) ([]Content, error) {
	if len(ids) == 0 {
		return []Content{}, nil
	}
	var contents []Content
	err := r.db.NewSelect().Model(&contents).Where("content_id IN (?)", bun.In(ids)).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.FindByIds: %w", err)
	}
	return contents, nil
}

func (r *ContentRepositoryImpl) FindByFilter(filter map[string]interface{}) ([]Content, error) {
	var contents []Content
	q := r.db.NewSelect().Model(&contents)
	for col, val := range filter {
		q = q.Where("? = ?", bun.Ident(col), val)
	}
	err := q.Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ContentRepository.FindByFilter: %w", err)
	}
	return contents, nil
}

func (r *ContentRepositoryImpl) Delete(id int64) error {
	_, err := r.db.NewDelete().Model((*Content)(nil)).Where("content_id = ?", id).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("ContentRepository.Delete: %w", err)
	}
	return nil
}
