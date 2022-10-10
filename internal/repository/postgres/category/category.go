package category

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

type Search interface {
	CategoryGetById(context.Context, int) (*model.Category, error)
	CategoryGetByTitle(context.Context, string) (*model.Category, error)
}

const (
	categoryTable = "category"
)

var (
	NotFoundError = errors.New("category not found")
)

type Category struct {
	mutex *sync.RWMutex
	db    *sqlx.DB
}

func NewCategory(db *sqlx.DB) *Category {
	return &Category{
		mutex: &sync.RWMutex{},
		db:    db,
	}
}

func (s Category) Categories(ctx context.Context) (cs []model.Category, err error) {
	query := fmt.Sprintf(`SELECT id, title FROM %s`, categoryTable)
	if err = s.db.SelectContext(ctx, &cs, query); err != nil {
		s.mutex.RUnlock()
		return nil, errors.Wrap(err, "all categories")
	}

	return
}

func (s *Category) AddCategory(ctx context.Context, title string) ([]model.Category, error) {
	s.mutex.Lock()

	_, err := s.CategoryGetByTitle(ctx, title)
	if !errors.Is(err, NotFoundError) {
		s.mutex.Unlock()
		return s.Categories(ctx)
	}

	var categoryId int
	createCategoryQuery := fmt.Sprintf("INSERT INTO %s (title) values ($1) RETURNING id", categoryTable)
	row := s.db.QueryRowContext(ctx, createCategoryQuery, title)
	err = row.Scan(&categoryId)
	if err != nil {
		s.mutex.Unlock()
		return nil, errors.Wrap(err, "insert category")
	}

	s.mutex.Unlock()

	return s.Categories(ctx)
}

func (s *Category) DeleteCategory(ctx context.Context, id int) ([]model.Category, error) {
	s.mutex.Lock()
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, categoryTable)
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		s.mutex.Unlock()
		return nil, errors.Wrap(err, "delete category")
	}
	s.mutex.Unlock()

	return s.Categories(ctx)
}

func (s *Category) CategoryGetById(ctx context.Context, id int) (cat *model.Category, err error) {
	var c model.Category
	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE id = %d`, categoryTable, id)
	if err = s.db.GetContext(ctx, &c, query); err != nil {
		return nil, errors.Wrap(NotFoundError, fmt.Sprintf("category '%d' not found", id))
	}

	return &c, nil
}

func (s *Category) CategoryGetByTitle(ctx context.Context, title string) (cat *model.Category, err error) {
	var c model.Category
	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE title = '%s'`, categoryTable, title)
	if err = s.db.GetContext(ctx, &c, query); err != nil {
		return nil, errors.Wrap(NotFoundError, fmt.Sprintf("category '%s' not found", title))
	}

	return &c, nil
}
