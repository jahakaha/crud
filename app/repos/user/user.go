package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Name        string
	Data        string
	Permissions int
}

type UserStore interface {
	Create(ctx context.Context, u User) (*uuid.UUID, error)
	Read(ctx context.Context, uid uuid.UUID) (*User, error)
	Delete(ctx context.Context, uid uuid.UUID) error
	SearchUser(ctx context.Context, name string) (chan User, error)
}

type Users struct {
	ustore UserStore
}

func NewUsers(ustore UserStore) *Users {
	return &Users{
		ustore: ustore,
	}
}

func (us *Users) Create(ctx context.Context, u User) (*User, error) {
	id, err := us.ustore.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("create user error:  %w", err)
	}
	u.ID = *id
	return &u, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*User, error) {
	u, err := us.ustore.Read(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("Read user error:  %w", err)
	}
	return u, nil
}

func (us *Users) Delete(ctx context.Context, uid uuid.UUID) (*User, error) {
	user, err := us.ustore.Read(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("Search user error:  %w", err)
	}
	return user, us.ustore.Delete(ctx, uid)
}

func (us *Users) SearchUsers(ctx context.Context, name string) (chan User, error) {

	// FIX ME need to use pattern Unit of Work
	// Business transaction

	chin, err := us.ustore.SearchUser(ctx, name)
	if err != nil {
		return nil, err
	}
	chout := make(chan User, 100)
	go func() {
		defer close(chout)
		for {
			select {
			case <-ctx.Done():
				return
			case u, ok := <-chin:
				if !ok {
					return
				}
				u.Permissions = 0755
				chout <- u
			}

		}
	}()
	return chout, nil
}
