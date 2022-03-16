package server

import (
	"context"
	"crud/app/repos/user"
	"net/http"
	"time"
)

type Server struct {
	srv http.Server
	us  *user.Users
}

func NewServer(addr string, h http.Handler) *Server {
	s := &Server{}
	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	s.srv.Shutdown(ctx)
	cancel()
}

func (s *Server) Start(us *user.Users) {
	s.us = us
	go s.srv.ListenAndServe()
}
