package main

import (
	"log"
	"net/http"
	"os"

	"github.com/LDmitryLD/url-shortener/internal/config"
	"github.com/LDmitryLD/url-shortener/internal/http_server/handlers/redirect"
	"github.com/LDmitryLD/url-shortener/internal/http_server/handlers/url/save"
	mwLogger "github.com/LDmitryLD/url-shortener/internal/http_server/middleware/logger"
	"github.com/LDmitryLD/url-shortener/internal/infrastructure/logs"
	"github.com/LDmitryLD/url-shortener/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("err load env: ", err)
	}

	cfg := config.MustLoad()

	logger := logs.NewLogger(*cfg, os.Stdout)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		logger.Error("failed to init storage: ", zap.Error(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	//router.Use(middleware.Logger)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shrtener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(logger, storage))
	})

	router.Get("/{alias}", redirect.New(logger, storage))

	logger.Info("starting server", zap.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server")
	}

	logger.Error("server stopped")
}
