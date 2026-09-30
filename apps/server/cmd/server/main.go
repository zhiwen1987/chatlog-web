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
	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
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

	// 加载部署许可 claims（R42.8）。失败不崩溃：nil=默认拒绝，记录原因。
	// 配置了 CHATLOG_LICENSE_VERIFY_KEY 走 JWS 验签（防篡改/伪造），否则保持直接解析兼容现状。
	var license *model.Claims
	func() {
		lctx, lcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer lcancel()
		var loaded *model.Claims
		var lerr error
		if len(cfg.LicenseVerifyKey) > 0 {
			loaded, lerr = db.LoadLicenseClaimsVerified(lctx, conn, cfg.DeploymentID, cfg.LicenseVerifyKey, cfg.LicenseExpectedAud)
			if lerr != nil {
				log.Printf("license claims verify failed (default deny): %v", lerr)
				return
			}
		} else {
			loaded, lerr = db.LoadLicenseClaims(lctx, conn, cfg.DeploymentID)
			if lerr != nil {
				log.Printf("license claims load failed (default deny): %v", lerr)
				return
			}
		}
		license = loaded
	}()
	if license != nil {
		log.Printf("license claims loaded for deployment %s (revision %d)", cfg.DeploymentID, license.LicenseRevision)
	}

	srv := handler.New(handler.Deps{
		DB:                  conn,
		JWTSecret:           cfg.JWTSecret,
		TokenTTLMin:         cfg.TokenTTLMinutes,
		License:             license,
		DeploymentID:        cfg.DeploymentID,
		LicenseVerifyKey:    cfg.LicenseVerifyKey,
		LicenseExpectedAud:  cfg.LicenseExpectedAud,
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
