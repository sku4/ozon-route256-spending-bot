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
	if u, err = us.GetById(ctx, telegramId); err == nil {
		return u, nil
	}

	mutex.Lock()
	defer mutex.Unlock()

	st, err := us.reposState.AddState(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "add user")
	}

	var userId int
	row := us.db.QueryRowContext(ctx, queryInsert, telegramId, st.Id)
	err = row.Scan(&userId)
	if err != nil {
		return nil, errors.Wrap(err, "insert user")
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

func (us *Users) GetById(ctx context.Context, id int) (u *User, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	var user model.UserDB
	if err = us.db.GetContext(ctx, &user, queryGetByTelegramId, id); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("user '%d' not found", id))
	}

	st, err := us.reposState.GetById(ctx, user.StateId)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", id))
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
