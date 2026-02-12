package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/yiannis54/go-socket-server/internal/config"
	"github.com/yiannis54/go-socket-server/internal/middleware"
	"github.com/yiannis54/go-socket-server/internal/notifications"
	"github.com/yiannis54/go-socket-server/internal/sockets"
)

const shutdownTimeout = 10 * time.Second

func Run(cfg *config.EnvConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	defer stop()

	socketHub := sockets.NewHub()
	notificationsClient := notifications.NewClient(socketHub)

	router, err := initRoutes(socketHub, cfg)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		ReadHeaderTimeout: 3 * time.Second, //nolint:mnd
		Handler:           router,
	}

	// errgroup: if any goroutine returns an error, the derived context is
	// cancelled, which causes all the others to shut down as well.
	g, ctx := errgroup.WithContext(ctx)

	// Socket hub
	g.Go(func() error {
		socketHub.Run(ctx)
		return nil
	})

	// gRPC server
	g.Go(func() error {
		return runRpc(ctx, notificationsClient, cfg)
	})

	// HTTP server
	g.Go(func() error {
		log.Printf("Listening Socket server on :%v", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	// Graceful shutdown watcher: waits for context cancellation (signal or
	// goroutine failure), then shuts down the HTTP server with a timeout.
	g.Go(func() error {
		<-ctx.Done()
		log.Println("Shutting down servers")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	})

	return g.Wait()
}

func initRoutes(socketHub *sockets.Hub, cfg *config.EnvConfig) (http.Handler, error) {
	mux := http.NewServeMux()
	wsHandler := middleware.AuthMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sockets.ServeWs(socketHub, w, r)
	}))
	mux.Handle("GET /ws", wsHandler)
	return mux, nil
}
