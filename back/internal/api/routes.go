package api

import (
	"net/http"

	"github.com/build-assistant/back/internal/api/handlers"
	"github.com/build-assistant/back/internal/api/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(h *handlers.Handlers) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS())

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", h.Auth.Login)
		r.Get("/callback", h.Auth.Callback)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/events", func(r chi.Router) {
			r.Use(middleware.TokenAuth(h.Events.TokenValidator))
			r.Post("/", h.Events.HandleEvent)
		})

		r.Route("/artifacts", func(r chi.Router) {
			r.Use(middleware.TokenAuth(h.Artifacts.TokenValidator))
			r.Post("/presign", h.Artifacts.Presign)
			r.Post("/confirm", h.Artifacts.ConfirmUpload)
			r.Post("/container", h.Artifacts.RegisterContainer)
			r.Delete("/{id}", h.Artifacts.DeleteArtifact)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.OIDCAuth(h.OIDCService, h.TokenCache))
			r.Get("/userinfo", h.Auth.UserInfo)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.OIDCAuth(h.OIDCService, h.TokenCache))
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", h.Projects.List)
				r.Post("/", h.Projects.Create)
				r.Get("/{name}", h.Projects.Get)
				r.Patch("/{name}", h.Projects.Update)
				r.Delete("/{name}", h.Projects.Delete)
				r.Post("/{name}/test-telegram", h.Projects.TestTelegramNotification)
				r.Post("/{name}/test-webhook", h.Projects.TestWebhook)
				r.Post("/{name}/test-b24", h.Projects.TestB24Notification)
				r.Route("/{project_name}/branches", func(r chi.Router) {
					r.Get("/", h.Branches.List)
					r.Post("/", h.Branches.Create)
					r.Get("/{branch_name}", h.Branches.Get)
					r.Patch("/{branch_name}", h.Branches.Update)
					r.Get("/{branch_name}/builds", h.Branches.GetBuilds)
					r.Post("/{branch_name}/redeploy", h.Branches.Redeploy)
					r.Post("/{branch_name}/test-telegram", h.Branches.TestTelegramNotification)
					r.Post("/{branch_name}/test-webhook", h.Branches.TestWebhook)
					r.Post("/{branch_name}/test-b24", h.Branches.TestB24Notification)
				})
			})

			r.Route("/builds", func(r chi.Router) {
				r.Get("/", h.Builds.List)
				r.Get("/{id}", h.Builds.Get)
				r.Get("/{id}/artifacts", h.Artifacts.GetBuildArtifacts)
				r.Delete("/{id}/artifacts", h.Artifacts.DeleteBuildArtifacts)
			})

			r.Route("/tokens", func(r chi.Router) {
				r.Get("/", h.Tokens.List)
				r.Post("/", h.Tokens.Create)
				r.Delete("/{id}", h.Tokens.Delete)
			})

			r.Route("/ssh-keys", func(r chi.Router) {
				r.Get("/", h.SSHKeys.List)
				r.Post("/", h.SSHKeys.Create)
				r.Delete("/{id}", h.SSHKeys.Delete)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return r
}
