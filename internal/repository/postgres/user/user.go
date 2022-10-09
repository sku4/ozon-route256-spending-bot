package user

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/memory/state"
	"sync"
)

var (
	users Users
	mutex sync.RWMutex
)

type Users []*User

type User struct {
	id    int
	state *state.State
}

func (u *User) GetState() *state.State {
	mutex.RLock()
	defer mutex.RUnlock()

	return u.state
}

func NewUsers() *Users {
	users = make(Users, 0)

	return &users
}

func (us *Users) AddUser(id int) (u *User, err error) {
	if u, err := us.GetUserById(id); err == nil {
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
	users = append(users, u)

	return u, nil
}

func (us *Users) GetUserById(id int) (u *User, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, user := range users {
		if user.id == id {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}
