package user

import (
	"context"
	"slices"
)

type InMemoryStore struct {
	users []User
}

func (s *InMemoryStore) Create(ctx context.Context, userFC User) (User, error) {
	s.users = append(s.users, userFC)
	return userFC, nil
}

func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	var deleted bool

	s.users = slices.DeleteFunc(s.users, func(user User) bool {
		deleted = true
		return user.ID == id
	})

	if !deleted {
		return &NotFoundError{ID: id}
	}

	return nil
}

func (s *InMemoryStore) Update(ctx context.Context, userFU User) error {
	var found bool

	for i, el := range s.users {
		if el.ID == userFU.ID {
			found = true
			s.users[i] = userFU
		}
	}

	if !found {
		return &NotFoundError{ID: userFU.ID}
	}
	return nil
}

func (s *InMemoryStore) FindOneByID(ctx context.Context, id string) (User, error) {
	for _, el := range s.users {
		if el.ID == id {
			return el, nil
		}
	}
	return User{}, &NotFoundError{ID: id}
}

func (s *InMemoryStore) FindOneByEmail(ctx context.Context, email string) (User, error) {
	for _, el := range s.users {
		if el.Email == email {
			return el, nil
		}
	}
	return User{}, &NotFoundError{Email: email}
}

func (s *InMemoryStore) FindAll(ctx context.Context) ([]User, error) {
	return s.users, nil
}
