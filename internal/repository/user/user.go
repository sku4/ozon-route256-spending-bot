package user

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/state"
	"sync"
)

var (
	users Users
	mutex sync.RWMutex
)

type Users []*User

type User struct {
	Id    int
	State *state.State
}

func NewUsers() *Users {
	users = make(Users, 0, 0)

	return &users
}

func (us *Users) AddUser(id int) *User {
	if u, err := us.GetUserById(id); err == nil {
		return u
	}

	mutex.Lock()
	defer mutex.Unlock()
	u := &User{
		Id:    id,
		State: state.NewState(),
	}
	users = append(users, u)

	return u
}

func (us *Users) GetUserById(id int) (u *User, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, user := range users {
		if user.Id == id {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}
