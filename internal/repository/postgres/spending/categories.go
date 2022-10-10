package spending

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
)

type CategorySearch interface {
	CategoryGetById(int) (*model.Category, error)
	CategoryGetByTitle(string) (*model.Category, error)
}

var (
	categoryNotFoundError = errors.New("category not found")
)

func (s Spending) Categories(context.Context) (cs []model.Category, err error) {
	s.mutex.RLock()

	query := fmt.Sprintf(`SELECT id, title FROM %s`, categoryTable)
	if err = s.db.Select(&cs, query); err != nil {
		s.mutex.RUnlock()
		return nil, errors.Wrap(err, "all categories")
	}

	s.mutex.RUnlock()

	return
}

func (s *Spending) AddCategory(ctx context.Context, title string) ([]model.Category, error) {
	s.mutex.Lock()

	_, err := s.CategoryGetByTitle(title)
	if !errors.Is(err, categoryNotFoundError) {
		s.mutex.Unlock()
		return s.Categories(ctx)
	}

	var categoryId int
	createCategoryQuery := fmt.Sprintf("INSERT INTO %s (title) values ($1) RETURNING id", categoryTable)
	row := s.db.QueryRow(createCategoryQuery, title)
	err = row.Scan(&categoryId)
	if err != nil {
		s.mutex.Unlock()
		return nil, errors.Wrap(err, "insert category")
	}

	s.mutex.Unlock()

	return s.Categories(ctx)
}

func (s *Spending) DeleteCategory(ctx context.Context, id int) ([]model.Category, error) {
	s.mutex.Lock()
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, categoryTable)
	_, err := s.db.Exec(query, id)
	if err != nil {
		s.mutex.Unlock()
		return nil, errors.Wrap(err, "delete category")
	}
	s.mutex.Unlock()

	return s.Categories(ctx)
}

func (s *Spending) CategoryGetById(id int) (cat *model.Category, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE id = %d`, categoryTable, id)
	if err = s.db.Select(&cat, query); err != nil {
		return nil, errors.Wrap(categoryNotFoundError, fmt.Sprintf("category '%d' not found", id))
	}

	return
}

func (s *Spending) CategoryGetByTitle(title string) (cat *model.Category, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := fmt.Sprintf(`SELECT id, title FROM %s WHERE title = "%s"`, categoryTable, title)
	if err = s.db.Select(&cat, query); err != nil {
		return nil, errors.Wrap(categoryNotFoundError, fmt.Sprintf("category '%s' not found", title))
	}

	return
}
