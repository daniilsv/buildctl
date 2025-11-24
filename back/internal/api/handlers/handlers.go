package handlers

import (
	authhandler "github.com/build-assistant/back/internal/api/handlers/auth"
	"github.com/build-assistant/back/internal/api/handlers/artifacts"
	"github.com/build-assistant/back/internal/api/handlers/branches"
	"github.com/build-assistant/back/internal/api/handlers/builds"
	"github.com/build-assistant/back/internal/api/handlers/events"
	"github.com/build-assistant/back/internal/api/handlers/projects"
	"github.com/build-assistant/back/internal/api/handlers/tokens"
	"github.com/build-assistant/back/internal/auth"
	"github.com/build-assistant/back/internal/notifications"
)

type Handlers struct {
	Auth        *authhandler.Handler
	Projects    *projects.Handler
	Branches    *branches.Handler
	Builds      *builds.Handler
	Events      *events.Handler
	Artifacts   *artifacts.Handler
	Tokens      *tokens.Handler
	OIDCService *auth.OIDCService
	TokenCache  *auth.TokenCache
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{
		Auth:        authhandler.NewHandler(deps.AuthService),
		Projects:    projects.NewHandler(deps.ProjectService, deps.Notifier),
		Branches:    branches.NewHandler(deps.BranchService, deps.Notifier),
		Builds:      builds.NewHandler(deps.BuildService),
		Events:      events.NewHandler(deps.EventService, deps.TokenValidator),
		Artifacts:   artifacts.NewHandler(deps.ArtifactService, deps.BuildService, deps.TokenValidator),
		Tokens:      tokens.NewHandler(deps.TokenService),
		OIDCService: deps.OIDCService,
		TokenCache:  deps.TokenCache,
	}
}

type Dependencies struct {
	AuthService      authhandler.Service
	ProjectService   projects.Service
	BranchService    branches.Service
	BuildService     builds.Service
	EventService     events.Service
	ArtifactService  artifacts.Service
	TokenService     tokens.Service
	TokenValidator   tokens.TokenValidator
	Notifier         notifications.Notifier
	OIDCService      *auth.OIDCService
	TokenCache       *auth.TokenCache
}

