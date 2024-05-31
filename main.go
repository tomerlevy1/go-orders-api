package main

import (
	"context"
	"fmt"

	"github.com/tomerlevy1/go-orders-api/application"
)

func main() {
	app := application.New()

	err := app.Start(context.TODO())
	if err != nil {
		// what's the difference?
		fmt.Errorf("failed to start app %w", err)
		fmt.Println("failed to start app", err)
	}
}
