package handlers

import (
	"github.com/build-assistant/back/internal/api/handlers/artifacts"
	"github.com/build-assistant/back/internal/api/handlers/auth"
	"github.com/build-assistant/back/internal/api/handlers/branches"
	"github.com/build-assistant/back/internal/api/handlers/builds"
	"github.com/build-assistant/back/internal/api/handlers/events"
	"github.com/build-assistant/back/internal/api/handlers/projects"
	"github.com/build-assistant/back/internal/api/handlers/tokens"
)

type Handlers struct {
	Auth      *auth.Handler
	Projects  *projects.Handler
	Branches  *branches.Handler
	Builds    *builds.Handler
	Events    *events.Handler
	Artifacts *artifacts.Handler
	Tokens    *tokens.Handler
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{
		Auth:      auth.NewHandler(deps.AuthService),
		Projects:  projects.NewHandler(deps.ProjectService),
		Branches:  branches.NewHandler(deps.BranchService),
		Builds:    builds.NewHandler(deps.BuildService),
		Events:    events.NewHandler(deps.EventService, deps.TokenValidator),
		Artifacts: artifacts.NewHandler(deps.ArtifactService, deps.TokenValidator),
		Tokens:    tokens.NewHandler(deps.TokenService),
	}
}

type Dependencies struct {
	AuthService      auth.Service
	ProjectService   projects.Service
	BranchService    branches.Service
	BuildService     builds.Service
	EventService     events.Service
	ArtifactService  artifacts.Service
	TokenService     tokens.Service
	TokenValidator   tokens.TokenValidator
}

