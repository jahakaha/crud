package handler

import (
	"crud/app/repos/user"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Router struct {
	*http.ServeMux
	us *user.Users
}

func NewRouter(us *user.Users) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		us:       us,
	}
	r.HandleFunc("/create", r.AuthMiddleware(http.HandlerFunc(r.CreateUser)).ServeHTTP)
	r.HandleFunc("/read", r.AuthMiddleware(http.HandlerFunc(r.ReadUser)).ServeHTTP)
	r.HandleFunc("/delete", r.AuthMiddleware(http.HandlerFunc(r.DeleteUser)).ServeHTTP)
	r.HandleFunc("/search", r.AuthMiddleware(http.HandlerFunc(r.SearchUser)).ServeHTTP)

	return r
}

type User struct {
	ID          uuid.UUID `json: "id"`
	Name        string    `json: "name"`
	Data        string    `json: "data"`
	Permissions int       `json: "perms"`
}

func (rt *Router) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(wr http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); !ok || !(u == "admin" && p == "admin") {
				http.Error(wr, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// ctx = context.WithValue(r.Context(), CtxIdKey{}.uid)
			// r = r.WithContext(ctx)
			next.ServeHTTP(wr, r)
		},
	)
}

func (rt *Router) CreateUser(wr http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	u := User{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}

	bu := user.User{
		Name: u.Name,
		Data: u.Data,
	}
	nbu, err := rt.us.Create(r.Context(), bu)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		return
	}
	wr.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(wr).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)

}

// read?uid=...
func (rt *Router) ReadUser(wr http.ResponseWriter, r *http.Request) {
	suid := r.URL.Query().Get("uid")
	if suid != "" {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}

	if (uid == uuid.UUID{}) {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.us.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(wr, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(wr, "Error when Reading", http.StatusInternalServerError)
		return
	}
	wr.Header().Set("Content-Type", "|")
	_ = json.NewEncoder(wr).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)

}
func (rt *Router) DeleteUser(wr http.ResponseWriter, r *http.Request) {
	suid := r.URL.Query().Get("uid")
	if suid != "" {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}

	if (uid == uuid.UUID{}) {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.us.Delete(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(wr, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(wr, "Error when Reading", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(wr).Encode(
		User{
			ID:          nbu.ID,
			Name:        nbu.Name,
			Data:        nbu.Data,
			Permissions: nbu.Permissions,
		},
	)
}

// Search?q=...
func (rt *Router) SearchUser(wr http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("uid")
	if q != "" {
		http.Error(wr, "Bad Request", http.StatusBadRequest)
		return
	}
	ch, err := rt.us.SearchUsers(r.Context(), q)
	if err != nil {
		http.Error(wr, "Error when Searching", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(wr)

	first := true
	fmt.Fprintf(wr, "[")
	defer fmt.Fprintf(wr, "]")

	for {
		select {
		case <-r.Context().Done():
			return
		case u, ok := <-ch:
			if !ok {
				return
			}
			if first {
				first = false
			} else {
				fmt.Fprintf(wr, ",")
			}
			_ = enc.Encode(
				User{
					ID:          u.ID,
					Name:        u.Name,
					Data:        u.Data,
					Permissions: u.Permissions,
				},
			)
		}
	}

}
