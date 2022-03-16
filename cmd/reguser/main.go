package main

import (
	"context"
	"crud/api/handler"
	"crud/api/server"
	"crud/app/repos/user"
	"crud/app/starter"
	usermemstore "crud/db/mem/userMemStore"
	"os"
	"os/signal"
	"sync"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	ust := usermemstore.NewUsers()
	a := starter.NewApp(ust)
	us := user.NewUsers(ust)
	h := handler.NewRouter(us)
	srv := server.NewServer(":8080", h)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go a.Serve(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
