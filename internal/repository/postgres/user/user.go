package user

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/state"
	"sync"
)

var (
	mutex = &sync.RWMutex{}
)

type Users struct {
	users []*User
	db    *sqlx.DB
}

type User struct {
	id    int
	state *state.State
}

func (u *User) GetState() *state.State {
	mutex.RLock()
	defer mutex.RUnlock()

	return u.state
}

func NewUsers(db *sqlx.DB) *Users {
	us := &Users{
		users: make([]*User, 0),
		db:    db,
	}

	return us
}

func (us *Users) AddUser(id int) (u *User, err error) {
	if u, err = us.GetUserById(id); err == nil {
		return u, nil
	}

	mutex.Lock()
	defer mutex.Unlock()

	s, err := state.NewState()
	if err != nil {
		return nil, errors.Wrap(err, "add user")
	}
	u = &User{
		id:    id,
		state: s,
	}
	us.users = append(us.users, u)

	return u, nil
}

func (us *Users) GetUserById(id int) (u *User, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, user := range us.users {
		if user.id == id {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}
