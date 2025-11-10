// Package main ...
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
	}()

	app, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("NewApp : %v", err)
	}

	go func() {
		if err := app.ListenAndServeGRPC(); err != nil {
			log.Fatalf("ListenAndServeGRPC : %v", err)
		}
	}()

	go func() {
		if err := app.ListenAndServeHTTP(); err != nil {
			log.Fatalf("ListenAndServeHTTP : %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	app.Close(ctx)

	log.Println("Server gracefully stopped")

}
