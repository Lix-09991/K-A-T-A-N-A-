package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Lix-09991/neon-saga-api/internal/config"
	"github.com/Lix-09991/neon-saga-api/internal/handler"
	custommw "github.com/Lix-09991/neon-saga-api/internal/middleware"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

func New(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// --- Глобальные мидлвары ---
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(custommw.Logger)
	r.Use(middleware.Recoverer)

	// --- CORS ---
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL, "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- DI: репозитории ---
	userRepo := repository.NewUserRepository(pool)

	// --- DI: сервисы ---
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresIn)

	// --- DI: хендлеры ---
	authHandler := handler.NewAuthHandler(authService)

	// --- Health ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	// --- API v1 ---
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Neon Saga API v1"))
		})

		// Auth (публичные)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Защищённые роуты
		r.Group(func(r chi.Router) {
			r.Use(custommw.RequireAuth([]byte(cfg.JWTSecret)))
			// здесь появятся /me, /users, /anime и т.д.
		})
	})

	return r
}
