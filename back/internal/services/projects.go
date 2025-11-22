package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/projects"
)

type ProjectService struct {
	queries *db.Queries
}

func NewProjectService(queries *db.Queries) *ProjectService {
	return &ProjectService{queries: queries}
}

func (s *ProjectService) CreateProject(ctx context.Context, req projects.CreateProjectRequest) (*projects.Project, error) {
	settingsJSON, err := json.Marshal(req.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	title := req.Name
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	}

	dbProject, err := s.queries.CreateProject(ctx, &db.CreateProjectParams{
		Name:           req.Name,
		Title:          title,
		RepositoryUrl:  req.RepositoryURL,
		RepositoryType: req.RepositoryType,
		AccessToken:    req.AccessToken,
		Settings:       settingsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return toProject(dbProject), nil
}

func (s *ProjectService) GetProject(ctx context.Context, name string) (*projects.Project, error) {
	dbProject, err := s.queries.GetProjectByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return toProject(dbProject), nil
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]projects.Project, error) {
	dbProjects, err := s.queries.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]projects.Project, len(dbProjects))
	for i, p := range dbProjects {
		result[i] = *toProject(p)
	}

	return result, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, name string, req projects.UpdateProjectRequest) (*projects.Project, error) {
	existing, err := s.queries.GetProjectByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	projectName := existing.Name
	if req.Name != nil {
		projectName = *req.Name
	}

	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}

	repoURL := existing.RepositoryUrl
	if req.RepositoryURL != nil {
		repoURL = *req.RepositoryURL
	}

	repoType := existing.RepositoryType
	if req.RepositoryType != nil {
		repoType = *req.RepositoryType
	}

	accessToken := existing.AccessToken
	if req.AccessToken != nil {
		accessToken = *req.AccessToken
	}

	settings := existing.Settings
	if req.Settings != nil {
		settingsJSON, err := json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
		settings = settingsJSON
	}

	dbProject, err := s.queries.UpdateProject(ctx, &db.UpdateProjectParams{
		ID:             existing.ID,
		Name:           projectName,
		Title:          title,
		RepositoryUrl:  repoURL,
		RepositoryType: repoType,
		AccessToken:    accessToken,
		Settings:       settings,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return toProject(dbProject), nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, name string) error {
	project, err := s.queries.GetProjectByName(ctx, name)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	return s.queries.DeleteProject(ctx, project.ID)
}

func (s *ProjectService) GetProjectByName(ctx context.Context, name string) (*projects.Project, error) {
	dbProject, err := s.queries.GetProjectByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return toProject(dbProject), nil
}

func toProject(p db.Project) *projects.Project {
	var settings map[string]interface{}
	if len(p.Settings) > 0 {
		json.Unmarshal(p.Settings, &settings)
	}

	return &projects.Project{
		ID:             p.ID.String(),
		Name:           p.Name,
		Title:          p.Title,
		RepositoryURL:  p.RepositoryUrl,
		RepositoryType: p.RepositoryType,
		Settings:       settings,
		CreatedAt:      p.CreatedAt.Time.Format(time.RFC3339),
	}
}
