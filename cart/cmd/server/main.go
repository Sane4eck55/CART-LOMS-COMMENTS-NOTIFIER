// Package main ...
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/app"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		_ = logger.Sync()
	}()

	app, err := app.NewApp(ctx)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("NewApp : %v", err))
	}

	go func() {
		if err := app.ListenAndServe(); err != nil {
			logger.Fatalw(fmt.Sprintf("ListenAndServe : %v", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Infow("Shutting down server...")

	if err := app.Close(ctx); err != nil {
		logger.Fatalw(fmt.Sprintf("app.Close():%+v", err))
	}

	logger.Infow("Server gracefully stopped")

}
