package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	cfg    Config
}

func New(cfg Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisAddr,
		}),
		cfg: cfg,
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.ServerPort),
		Handler: a.router,
	}

	// We can use err := to declare and initialize the variable in the same line
	var err error

	err = a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Errorf("failed to close redis connection: %w", err)
		}
	}()

	fmt.Println("Starting server on port", a.cfg.ServerPort)

	ch := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}

		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = server.Shutdown(timeoutCtx)
		defer cancel()

		return server.Shutdown(timeoutCtx)
	}
	return nil
}
