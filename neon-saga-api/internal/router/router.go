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
	"github.com/Lix-09991/neon-saga-api/internal/models"
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

	// ============================================================
	// DI: репозитории
	// ============================================================
	userRepo := repository.NewUserRepository(pool)
	animeRepo := repository.NewAnimeRepository(pool)
	epRepo := repository.NewEpisodeRepository(pool)
	markerRepo := repository.NewMarkerRepository(pool)
	franchiseRepo := repository.NewFranchiseRepository(pool)
	collectionRepo := repository.NewCollectionRepository(pool)

	// ============================================================
	// DI: сервисы
	// ============================================================
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresIn)
	animeService := service.NewAnimeService(animeRepo, epRepo)
	episodeService := service.NewEpisodeService(epRepo, markerRepo, animeRepo)
	franchiseService := service.NewFranchiseService(franchiseRepo)
	collectionService := service.NewCollectionService(collectionRepo)

	// ============================================================
	// DI: хендлеры
	// ============================================================
	authHandler := handler.NewAuthHandler(authService)
	animeHandler := handler.NewAnimeHandler(animeService)
	episodeHandler := handler.NewEpisodeHandler(episodeService)
	franchiseHandler := handler.NewFranchiseHandler(franchiseService)
	collectionHandler := handler.NewCollectionHandler(collectionService)
	uploadHandler := handler.NewUploadHandler()

	// --- Health ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	// --- Статика: /static/videos/..., /static/images/... ---
	fs := http.FileServer(http.Dir("./uploads"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// ============================================================
	// API v1
	// ============================================================
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Neon Saga API v1"))
		})

		// ============================================================
		// ПУБЛИЧНЫЕ
		// ============================================================
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Anime
		r.Get("/anime", animeHandler.List)
		r.Get("/anime/facets", animeHandler.Facets)
		r.Get("/anime/{id}", animeHandler.Get)
		r.Get("/anime/{animeID}/episodes", episodeHandler.ListByAnime)

		// Episodes
		r.Get("/episodes/{id}", episodeHandler.Get)

		// Franchises
		r.Get("/franchises", franchiseHandler.List)
		r.Get("/franchises/{id}", franchiseHandler.Get)

		// Collections
		r.Get("/collections", collectionHandler.List)
		r.Get("/collections/{id}", collectionHandler.Get)

		// ============================================================
		// ЗАЩИЩЁННЫЕ (JWT)
		// ============================================================
		r.Group(func(r chi.Router) {
			r.Use(custommw.RequireAuth([]byte(cfg.JWTSecret)))

			// ============================================================
			// ADMIN / MODERATOR
			// ============================================================
			r.Group(func(r chi.Router) {
				r.Use(custommw.RequireRole(models.RoleAdmin, models.RoleModerator))

				// Anime CRUD
				r.Post("/anime", animeHandler.Create)
				r.Patch("/anime/{id}", animeHandler.Update)
				r.Delete("/anime/{id}", animeHandler.Delete)

				// Episodes
				r.Post("/anime/{animeID}/episodes", episodeHandler.Create)
				r.Delete("/episodes/{id}", episodeHandler.Delete)
				r.Put("/episodes/{id}/markers", episodeHandler.SetMarkers)

				// Franchises CRUD + items
				r.Post("/franchises", franchiseHandler.Create)
				r.Patch("/franchises/{id}", franchiseHandler.Update)
				r.Delete("/franchises/{id}", franchiseHandler.Delete)
				r.Put("/franchises/{id}/items", franchiseHandler.SetItems)

				// Collections CRUD + items
				r.Post("/collections", collectionHandler.Create)
				r.Patch("/collections/{id}", collectionHandler.Update)
				r.Delete("/collections/{id}", collectionHandler.Delete)
				r.Put("/collections/{id}/items", collectionHandler.SetItems)

				// Uploads
				r.Post("/admin/upload/video", uploadHandler.UploadVideo)
				r.Post("/admin/upload/image", uploadHandler.UploadImage)
			})
		})
	})

	return r
}
