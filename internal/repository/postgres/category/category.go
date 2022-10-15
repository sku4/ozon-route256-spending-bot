package category

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
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
	db *sqlx.DB
}

func NewCategory(db *sqlx.DB) *Category {
	return &Category{
		db: db,
	}
}

func (s Category) Categories(ctx context.Context) (cs []model.Category, err error) {
	query := fmt.Sprintf(`SELECT id, title FROM %s`, categoryTable)
	if err = s.db.SelectContext(ctx, &cs, query); err != nil {
		return nil, errors.Wrap(err, "all categories")
	}

	return
}

func (s *Category) AddCategory(ctx context.Context, title string) (categoryId int, err error) {
	_, err = s.CategoryGetByTitle(ctx, title)
	if !errors.Is(err, NotFoundError) {
		return 0, errors.Wrap(err, "add category")
	}

	createCategoryQuery := fmt.Sprintf("INSERT INTO %s (title) values ($1) RETURNING id", categoryTable)
	row := s.db.QueryRowContext(ctx, createCategoryQuery, title)
	err = row.Scan(&categoryId)
	if err != nil {
		return 0, errors.Wrap(err, "insert category")
	}

	return
}

func (s *Category) DeleteCategory(ctx context.Context, id int) (err error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, categoryTable)
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "delete category")
	}

	return
}

func (s *Category) CategoryGetById(ctx context.Context, id int) (cat *model.Category, err error) {
	var c model.Category
	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE id = %d`, categoryTable, id)
	if err = s.db.GetContext(ctx, &c, query); err != nil {
		return nil, NotFoundError
	}

	return &c, nil
}

func (s *Category) CategoryGetByTitle(ctx context.Context, title string) (cat *model.Category, err error) {
	var c model.Category
	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE title = '%s'`, categoryTable, title)
	if err = s.db.GetContext(ctx, &c, query); err != nil {
		return nil, NotFoundError
	}

	return &c, nil
}
