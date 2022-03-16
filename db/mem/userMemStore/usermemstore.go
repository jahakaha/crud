package usermemstore

import (
	"context"
	"crud/app/repos/user"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var _ user.UserStore = &Users{}

type Users struct {
	sync.Mutex
	m map[uuid.UUID]user.User
}

func NewUsers() *Users {
	return &Users{
		m: make(map[uuid.UUID]user.User),
	}
}

func (us *Users) Create(ctx context.Context, u user.User) (*uuid.UUID, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	uid := uuid.New()
	u.ID = uid
	us.m[u.ID] = u
	return &uid, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*user.User, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	u, ok := us.m[uid]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

// Did not return err if user is not exists
func (us *Users) Delete(ctx context.Context, uid uuid.UUID) error {
	us.Lock()
	us.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(us.m, uid)
	return nil

}

func (us *Users) SearchUser(ctx context.Context, str string) (chan user.User, error) {
	us.Lock()
	us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// FIX ME REWORK RADIX TREE

	chout := make(chan user.User, 100)
	go func() {
		defer close(chout)
		us.Lock()
		defer us.Unlock()
		for _, u := range us.m {
			if strings.Contains(u.Name, str) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					return
				case chout <- u:
				}
			}
		}
	}()

	return chout, nil
}
