package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/tomerlevy1/go-orders-api/application"
)

func main() {
	app := application.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Errorf("failed to start app %w", err)
	}
}
