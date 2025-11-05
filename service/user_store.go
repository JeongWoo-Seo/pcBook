package service

import "sync"

type UserStore interface {
	Save(user *User) error
	Find(username string) (*User, error)
}

type InmemoryUserStore struct {
	mutax sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() *InmemoryUserStore {
	return &InmemoryUserStore{
		users: make(map[string]*User),
	}
}

func (store *InmemoryUserStore) Save(user *User) error {
	store.mutax.Lock()
	defer store.mutax.Unlock()

	if store.users[user.Username] != nil {
		return ErrAlreadyExists
	}

	store.users[user.Username] = user.Clon()
	return nil
}

func (store *InmemoryUserStore) Find(username string) (*User, error) {
	store.mutax.Lock()
	defer store.mutax.Unlock()

	user := store.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clon(), nil
}
