package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/config"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/db"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/handler"
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/migration"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer conn.Close()

	// 等待数据库就绪（Docker Compose 启动竞态保护）
	waitForDB(conn, 30*time.Second)

	if err := migration.Migrate(context.Background(), conn); err != nil {
		log.Fatalf("migration: %v", err)
	}
	log.Println("migrations applied")

	srv := handler.New(handler.Deps{
		DB:          conn,
		JWTSecret:   cfg.JWTSecret,
		TokenTTLMin: cfg.TokenTTLMinutes,
	})

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("chatlog-server listening on %s", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}

func waitForDB(conn *sql.DB, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := conn.Ping(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	log.Fatalf("database not ready within %s", timeout)
}