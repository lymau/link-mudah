package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type config struct {
	Port         string
	DatabaseURL  string
	RedisAddress string
	JWTSecret    string
	UploadDir    string
}

func main() {
	cfg := loadConfig()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := connectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	rdb := connectRedis(cfg.RedisAddress)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	r := chi.NewRouter()
	r.Use(corsMiddleware)

	auth := newAuthHandler(db, cfg.JWTSecret)
	profile := newProfileHandler(db, rdb, cfg.UploadDir)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", auth.register)
		r.Post("/login", auth.login)
	})
	r.Route("/api/me", func(r chi.Router) {
		r.Use(RequireAuth(cfg.JWTSecret))
		r.Get("/", profile.me)
		r.Put("/settings", profile.updateSettings)
		r.Post("/avatar", profile.uploadAvatar)
		r.Delete("/avatar", profile.deleteAvatar)
		r.Post("/links", profile.createLink)
		r.Put("/links/{id}", profile.updateLink)
		r.Delete("/links/{id}", profile.deleteLink)
	})
	r.Get("/api/public/{username}", profile.publicPage)

	r.Get("/uploads/*", profile.serveUploads("/uploads"))
	r.Get("/api/uploads/*", profile.serveUploads("/api/uploads"))

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("failed to encode healthz response: %v", err)
		}
	}
	r.Get("/healthz", healthHandler)
	r.Get("/api/healthz", healthHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "linkmudah-backend",
			"healthz": "/healthz",
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func loadConfig() config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return config{
		Port:         port,
		DatabaseURL:  buildPostgresDSN(),
		RedisAddress: getEnv("REDIS_ADDR", "redis:6379"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		UploadDir:    getEnv("UPLOAD_DIR", "uploads"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildPostgresDSN() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}

	user := getEnv("POSTGRES_USER", "linkmudah")
	password := getEnv("POSTGRES_PASSWORD", "linkmudah_dev")
	host := getEnv("POSTGRES_HOST", "postgres")
	port := getEnv("POSTGRES_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "linkmudah")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
}

func connectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func connectRedis(address string) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: address})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	return client
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		reqHeaders := r.Header.Get("Access-Control-Request-Headers")
		if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, Origin")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

