package category_limit

import (
	"context"
	"database/sql"
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
	queryUpdate        = fmt.Sprintf(`UPDATE %s SET category_limit = $1 WHERE id = $2`, categoryLimitTable)
	queryInsert        = fmt.Sprintf(`INSERT INTO %s (state_id, category_id, category_limit) 
												values ($1, $2, $3) RETURNING id`, categoryLimitTable)
	querySelectByStateCategory = fmt.Sprintf(`SELECT id, state_id, category_id, category_limit 
									FROM %s WHERE state_id = $1 AND category_id = $2`, categoryLimitTable)
	querySelectByState = fmt.Sprintf(`SELECT id, state_id, category_id, category_limit 
									FROM %s WHERE state_id = $1`, categoryLimitTable)
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
	tx, err := cl.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "limit tx begin")
	}

	categoryLimitState, err := cl.TxGetByStateCategory(ctx, tx, stateId, categoryId)
	if err != nil && !errors.Is(err, limitNotFoundError) {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "limit rollback")
		}
		return nil, errors.Wrap(err, "limit set")
	} else if err == nil {
		// update limit
		_, err = tx.ExecContext(ctx, queryUpdate, limit.Original(), categoryLimitState.Id)
		if err != nil {
			errRoll := tx.Rollback()
			if errRoll != nil {
				return nil, errors.Wrap(errRoll, "limit rollback")
			}
			return nil, errors.Wrap(err, "limit update")
		}

		err = tx.Commit()
		if err != nil {
			return nil, errors.Wrap(err, "limit tx commit")
		}

		categoryLimitState.Limit = limit
		return categoryLimitState, nil
	}

	cat, err := cl.categorySearch.TxCategoryGetById(ctx, tx, categoryId)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "limit rollback")
		}
		return nil, errors.Wrap(err, "limit get")
	}

	var categoryLimitId int
	row := tx.QueryRowContext(ctx, queryInsert, stateId, cat.Id, limit.Original())
	err = row.Scan(&categoryLimitId)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "limit rollback")
		}
		return nil, errors.Wrap(err, "limit insert")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "limit tx commit")
	}

	return &CategoryLimit{
		CategoryLimit: model.CategoryLimit{
			Id:       categoryLimitId,
			Category: cat,
		},
		Limit: limit,
	}, nil
}

func (cl *CategoryLimit) TxGetByStateCategory(ctx context.Context, tx *sql.Tx, stateId, categoryId int) (c *CategoryLimit, err error) {
	var clDB model.CategoryLimitDB
	row := tx.QueryRowContext(ctx, querySelectByStateCategory, stateId, categoryId)
	err = row.Scan(&clDB.Id, &clDB.StateId, &clDB.CategoryId,
		&clDB.Limit)
	if err != nil {
		return nil, errors.Wrap(limitNotFoundError, "limit not found")
	}

	cat, err := cl.categorySearch.TxCategoryGetById(ctx, tx, clDB.CategoryId)
	if err != nil {
		return nil, errors.Wrap(err, "get by state category")
	}

	c = &CategoryLimit{
		CategoryLimit: model.CategoryLimit{
			Id:       clDB.Id,
			Category: cat,
		},
		Limit:          decimal.Decimal(clDB.Limit),
		db:             cl.db,
		categorySearch: cl.categorySearch,
	}

	return
}

func (cl *CategoryLimit) GetByState(ctx context.Context, stateId int) (cls []*CategoryLimit, err error) {
	var categoryLimitDB []*model.CategoryLimitDB
	if err = cl.db.SelectContext(ctx, &categoryLimitDB, querySelectByState, stateId); err != nil {
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
			Limit:          decimal.Decimal(limitDB.Limit),
			db:             cl.db,
			categorySearch: cl.categorySearch,
		}
		cls = append(cls, c)
	}

	return
}
