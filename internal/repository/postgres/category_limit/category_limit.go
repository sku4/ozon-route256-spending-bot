package category_limit

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
)

type CategoryLimitSet interface {
	Set(ctx context.Context, stateId, categoryId int, limit decimal.Decimal) (*CategoryLimit, error)
	GetByState(ctx context.Context, stateId int) ([]*CategoryLimit, error)
}

const (
	categoryLimitTable = "category_limit"
)

var (
	limitNotFoundError = errors.New("limit not found")
)

type CategoryLimit struct {
	model.CategoryLimit
	Limit          decimal.Decimal
	db             *sqlx.DB
	categorySearch category.Search
}

func NewCategoryLimit(db *sqlx.DB, categorySearch category.Search) *CategoryLimit {
	return &CategoryLimit{
		db:             db,
		categorySearch: categorySearch,
	}
}

func (cl *CategoryLimit) Set(ctx context.Context, stateId, categoryId int, limit decimal.Decimal) (s *CategoryLimit, err error) {
	categoryLimitState, err := cl.GetByStateCategory(ctx, stateId, categoryId)
	if err != nil && !errors.Is(err, limitNotFoundError) {
		return nil, errors.Wrap(err, "limit set")
	} else if err == nil {
		// update limit
		query := fmt.Sprintf(`UPDATE %s SET category_limit = $1 WHERE id = $2`, categoryLimitTable)
		_, err = cl.db.ExecContext(ctx, query, limit.Original(), categoryLimitState.Id)
		if err != nil {
			return nil, errors.Wrap(err, "limit update")
		}
		categoryLimitState.Limit = limit
		return categoryLimitState, nil
	}

	cat, err := cl.categorySearch.CategoryGetById(ctx, categoryId)
	if err != nil {
		return nil, errors.Wrap(err, "limit set")
	}

	var categoryLimitId int
	createLimitQuery := fmt.Sprintf(`INSERT INTO %s (state_id, category_id, category_limit) 
												values ($1, $2, $3) RETURNING id`, categoryLimitTable)
	row := cl.db.QueryRowContext(ctx, createLimitQuery, stateId, cat.Id, limit.Original())
	err = row.Scan(&categoryLimitId)
	if err != nil {
		return nil, errors.Wrap(err, "insert limit")
	}

	return &CategoryLimit{
		CategoryLimit: model.CategoryLimit{
			Id:       categoryLimitId,
			Category: cat,
		},
		Limit: limit,
	}, nil
}

func (cl *CategoryLimit) GetById(ctx context.Context, id int) (c *CategoryLimit, err error) {
	var categoryLimitDB model.CategoryLimitDB
	query := fmt.Sprintf(`SELECT id, state_id, category_id, category_limit 
									FROM %s WHERE id = %d`, categoryLimitTable, id)
	if err = cl.db.GetContext(ctx, &categoryLimitDB, query); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("limit '%d' not found", id))
	}

	cat, err := cl.categorySearch.CategoryGetById(ctx, categoryLimitDB.CategoryId)
	if err != nil {
		return nil, errors.Wrap(err, "get by id")
	}

	c = &CategoryLimit{
		CategoryLimit: model.CategoryLimit{
			Id:       categoryLimitDB.Id,
			Category: cat,
		},
		Limit:          decimal.ToDecimal(categoryLimitDB.Limit),
		db:             cl.db,
		categorySearch: cl.categorySearch,
	}

	return
}

func (cl *CategoryLimit) GetByStateCategory(ctx context.Context, stateId, categoryId int) (c *CategoryLimit, err error) {
	var categoryLimitDB model.CategoryLimitDB
	query := fmt.Sprintf(`SELECT id, state_id, category_id, category_limit 
									FROM %s WHERE state_id = $1 AND category_id = $2`, categoryLimitTable)
	if err = cl.db.GetContext(ctx, &categoryLimitDB, query, stateId, categoryId); err != nil {
		return nil, errors.Wrap(limitNotFoundError, "limit not found")
	}

	cat, err := cl.categorySearch.CategoryGetById(ctx, categoryLimitDB.CategoryId)
	if err != nil {
		return nil, errors.Wrap(err, "get by state category")
	}

	c = &CategoryLimit{
		CategoryLimit: model.CategoryLimit{
			Id:       categoryLimitDB.Id,
			Category: cat,
		},
		Limit:          decimal.ToDecimal(categoryLimitDB.Limit),
		db:             cl.db,
		categorySearch: cl.categorySearch,
	}

	return
}

func (cl *CategoryLimit) GetByState(ctx context.Context, stateId int) (cls []*CategoryLimit, err error) {
	var categoryLimitDB []*model.CategoryLimitDB
	query := fmt.Sprintf(`SELECT id, state_id, category_id, category_limit 
									FROM %s WHERE state_id = $1`, categoryLimitTable)
	if err = cl.db.SelectContext(ctx, &categoryLimitDB, query, stateId); err != nil {
		return nil, errors.Wrap(err, "get by state")
	}

	for _, limitDB := range categoryLimitDB {
		cat, err := cl.categorySearch.CategoryGetById(ctx, limitDB.CategoryId)
		if err != nil {
			return nil, errors.Wrap(err, "get by state")
		}

		c := &CategoryLimit{
			CategoryLimit: model.CategoryLimit{
				Id:       limitDB.Id,
				Category: cat,
			},
			Limit:          decimal.ToDecimal(limitDB.Limit),
			db:             cl.db,
			categorySearch: cl.categorySearch,
		}
		cls = append(cls, c)
	}

	return
}
