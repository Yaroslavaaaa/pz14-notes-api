// Package main Notes API server.
//
// @title           Notes API
// @version         1.0
// @description     Учебный REST API для заметок (CRUD).
// @contact.name    Backend Course
// @contact.email   example@university.ru
// @BasePath        /api/v1
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	httpx "example.com/notes-api/internal/http"
	"example.com/notes-api/internal/http/handlers"
	"example.com/notes-api/internal/repo"
)

func main() {
	// Загружаем переменные окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	log.Println("Connecting to DB:", dsn)

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(40) // максимум открытых соединений
	db.SetMaxIdleConns(25) // максимум соединений в простое
	db.SetConnMaxLifetime(5 * time.Minute)

	// Контекст с таймаутом для проверки соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	log.Println("Connected to DB successfully")

	// Инициализация репозитория PostgreSQL
	noteRepo := repo.NewNoteRepoPG(db)

	// HTTP handlers и роутер
	h := &handlers.Handler{Repo: noteRepo}
	r := httpx.NewRouter(h)

	// Swagger UI
	r.Get("/docs/*", httpSwagger.WrapHandler)
	r.Get("/docs/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// Запуск сервера
	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
