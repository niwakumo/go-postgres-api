package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"go-postgres-api/internal/handler"
	"go-postgres-api/internal/repository"
	"go-postgres-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx ドライバ登録
)

func main() {
	// 1. PostgreSQL 接続
	dsn := "postgres://postgres:postgres@localhost:5432/devdb?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// コネクションプールの設定
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 疎通確認
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	// 2. 手動DI (Pure DI)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	// 3. ルーティング設定
	r := chi.NewRouter()
	r.Use(middleware.Logger)    // リクエストログ出力
	r.Use(middleware.Recoverer) // panic時の500エラーハンドリング

	r.Get("/users/{id}", userHandler.GetUser)
	r.Post("/users", userHandler.CreateUser)

	// 4. サーバー起動
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}