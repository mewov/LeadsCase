package httpserver

import (
	handlers2 "demo-server/internal/httpserver/handlers"
	middleware2 "demo-server/internal/httpserver/middleware"
	"demo-server/internal/libs"
	"demo-server/internal/storage"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
)

func Register(router *chi.Mux, postgres *storage.Postgres, config *libs.Config, httpLogger *slog.Logger) {
	router.Use(middleware2.Logger(httpLogger))

	router.Get("/health", handlers2.HandleHealth)
	router.Route("/api", func(r chi.Router) {
		r.Post("/lead", handlers2.HandleCreateLead(postgres))
		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", handlers2.HandleAdminLogin(postgres, config.JWTSecret))
			r.Post("/refresh", handlers2.HandleAdminRefresh(postgres, config.JWTSecret))

			r.Group(func(r chi.Router) {
				r.Use(middleware2.Auth(config.JWTSecret))

				r.Get("/leads", handlers2.HandleListLeads(postgres))
				r.Get("/leads/{id}", handlers2.HandleGetLeadByID(postgres))
				r.Patch("/leads/{id}/status", handlers2.HandleChangeLeadStatus(postgres))
				r.Post("/logout", handlers2.HandleAdminLogout(postgres))
			})
		})
	})
}

func Listen(router *chi.Mux, config *libs.Config) error {
	server := &http.Server{
		Addr:         config.ServerAddr,
		Handler:      router,
		WriteTimeout: config.ServerWriteTimeout,
		ReadTimeout:  config.ServerReadTimeout,
	}

	return server.ListenAndServe()
}
