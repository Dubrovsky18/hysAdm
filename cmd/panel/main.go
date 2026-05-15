package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hysteria2-panel/internal/config"
	"hysteria2-panel/internal/database"
	"hysteria2-panel/internal/handlers"
	"hysteria2-panel/internal/middleware"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfgPath := "config.yaml"
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		cfgPath = env
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Printf("config not found at %s, using defaults", cfgPath)
		cfg = config.Default()
	}

	if env := os.Getenv("DATABASE_DSN"); env != "" {
		cfg.Database.DSN = env
	}
	if env := os.Getenv("JWT_SECRET"); env != "" {
		cfg.JWT.Secret = env
	}
	if env := os.Getenv("ADMIN_USERNAME"); env != "" {
		cfg.Panel.AdminUsername = env
	}
	if env := os.Getenv("ADMIN_PASSWORD"); env != "" {
		cfg.Panel.AdminPassword = env
	}
	if env := os.Getenv("PANEL_DOMAIN"); env != "" {
		cfg.Panel.Domain = env
	}

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.RunMigrations(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	seedAdmin(ctx, db, cfg)

	h := handlers.NewHandler(db.Pool, cfg.JWT.Secret, cfg.JWT.TTL, cfg.Panel.Domain)

	r := chi.NewRouter()
	r.Use(middleware.CORS())

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", h.Login)

		r.Get("/sub/{uuid}", h.GetSubscription)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWT.Secret))

			r.Get("/dashboard", h.GetDashboard)
			r.Get("/me", h.GetCurrentUser)

			r.Route("/users", func(r chi.Router) {
				r.Get("/", h.ListUsers)
				r.Post("/", h.CreateUser)
				r.Get("/traffic", h.GetAllUserTraffic)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", h.GetUser)
					r.Put("/", h.UpdateUser)
					r.Delete("/", h.DeleteUser)
					r.Post("/keys", h.CreateKey)
					r.Get("/keys", h.ListUserKeys)
					r.Post("/subscription", h.GenerateSubscription)
					r.Get("/subscriptions", h.GetSubscriptionHistory)
					r.Get("/traffic", h.GetUserTraffic)
				})
			})

			r.Route("/keys", func(r chi.Router) {
				r.Delete("/{keyId}", h.RevokeKey)
			})

			r.Route("/servers", func(r chi.Router) {
				r.Get("/", h.ListServers)
				r.Post("/", h.CreateServer)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", h.GetServer)
					r.Put("/", h.UpdateServer)
					r.Delete("/", h.DeleteServer)
					r.Put("/toggle", h.ToggleServer)
				})
			})

			r.Route("/domains", func(r chi.Router) {
				r.Get("/", h.ListDomains)
				r.Post("/", h.CreateDomain)
				r.Delete("/{id}", h.DeleteDomain)
			})

			r.Post("/traffic", h.RecordTraffic)
		})
	})

	// Serve admin frontend
	r.Handle("/*", http.FileServer(http.Dir("./web")))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Hysteria2 Admin Panel starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}

func seedAdmin(ctx context.Context, db *database.DB, cfg *config.Config) {
	var count int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE username = $1`, cfg.Panel.AdminUsername).Scan(&count)
	if err != nil {
		log.Printf("check admin: %v", err)
		return
	}

	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Panel.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("hash admin password: %v", err)
			return
		}

		b := make([]byte, 16)
		rand.Read(b)
		uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])

		_, err = db.Pool.Exec(ctx,
			`INSERT INTO users (uuid, username, password_hash, status, traffic_limit, email)
			 VALUES ($1, $2, $3, 'admin', 99999999999999, 'admin@panel.local')`,
			uuid, cfg.Panel.AdminUsername, string(hash),
		)
		if err != nil {
			log.Printf("seed admin failed: %v", err)
		} else {
			log.Printf("admin user '%s' created with password from config", cfg.Panel.AdminUsername)
		}
	}
}
