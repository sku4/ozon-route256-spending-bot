package user

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/state"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

const (
	userTable = "user"
)

var (
	mutex                = &sync.RWMutex{}
	queryGeById          = fmt.Sprintf(`SELECT id, state_id FROM "%s" WHERE id=$1`, userTable)
	queryInsert          = fmt.Sprintf(`INSERT INTO "%s" (telegram_id, state_id) values ($1, $2) RETURNING id`, userTable)
	queryGetByTelegramId = fmt.Sprintf(`SELECT id, state_id, telegram_id FROM "%s" WHERE telegram_id=$1`, userTable)
)

type Users struct {
	db         *sqlx.DB
	reposCurr  currency.Client
	reposState state.Client
}

type User struct {
	model.User
	State      *state.State
	db         *sqlx.DB
	reposCurr  currency.Client
	reposState state.Client
}

func (u *User) GetState(ctx context.Context) (s *state.State, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if u.State == nil {
		var user model.UserDB
		if err = u.db.GetContext(ctx, &user, queryGeById, u.TgId); err != nil {
			return nil, errors.Wrap(err, "user get state")
		}
		u.State, err = u.reposState.GetById(ctx, user.StateId)
		if err != nil {
			return nil, errors.Wrap(err, "user get state")
		}
	}

	return u.State, nil
}

func NewUsers(db *sqlx.DB, reposCurr currency.Client, reposState state.Client) *Users {
	us := &Users{
		db:         db,
		reposCurr:  reposCurr,
		reposState: reposState,
	}

	return us
}

func (us *Users) AddUser(ctx context.Context, telegramId int) (u *User, err error) {
	if u, err = us.GetByTgId(ctx, telegramId); err == nil {
		return u, nil
	}

	tx, err := us.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "add user tx begin")
	}

	st, err := us.reposState.AddStateTx(ctx, tx)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "add user rollback")
		}
		return nil, errors.Wrap(err, "add user")
	}

	var userId int
	row := tx.QueryRowContext(ctx, queryInsert, telegramId, st.Id)
	err = row.Scan(&userId)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "insert user rollback")
		}
		return nil, errors.Wrap(err, "insert user")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "add user tx commit")
	}

	u = &User{
		User: model.User{
			Id:   userId,
			TgId: telegramId,
		},
		State:      st,
		db:         us.db,
		reposCurr:  us.reposCurr,
		reposState: us.reposState,
	}

	return u, nil
}

func (us *Users) GetByTgId(ctx context.Context, tgId int) (u *User, err error) {
	tx, err := us.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "user by tg_id tx begin")
	}

	var user model.UserDB
	row := tx.QueryRowContext(ctx, queryGetByTelegramId, tgId)
	err = row.Scan(&user.Id, &user.StateId, &user.TgId)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "user by tg_id rollback")
		}
		return nil, errors.Wrap(err, fmt.Sprintf("user '%d' not found", tgId))
	}

	st, err := us.reposState.GetByIdTx(ctx, tx, user.StateId)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			return nil, errors.Wrap(errRoll, "user by tg_id rollback")
		}
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", tgId))
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "user by tg_id tx commit")
	}

	u = &User{
		User: model.User{
			Id:   user.Id,
			TgId: user.TgId,
		},
		State:      st,
		db:         us.db,
		reposCurr:  us.reposCurr,
		reposState: us.reposState,
	}

	return
}
