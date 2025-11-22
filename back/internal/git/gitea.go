package git

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "github.com/build-assistant/back/db/gen"
)

type GiteaClient struct {
	baseURL string
	client  *http.Client
}

func NewGiteaClient(baseURL string) *GiteaClient {
	return &GiteaClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GiteaClient) GetCommits(ctx context.Context, project *db.Project, branch *db.Branch, startHash, endHash string) ([]Commit, error) {
	var settings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse project settings: %w", err)
	}

	apiURL, ok := settings["git_api_url"].(string)
	if !ok {
		apiURL = c.baseURL
	}

	repo := project.RepositoryUrl
	url := fmt.Sprintf("%s/repos/%s/commits?sha=%s&limit=100", apiURL, repo, branch.Name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if project.AccessToken != "" {
		req.Header.Set("Authorization", "token "+project.AccessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("git API error: %s - %s", resp.Status, string(body))
	}

	var giteaCommits []struct {
		Sha    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&giteaCommits); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	commits := make([]Commit, 0)
	foundStart := false

	for _, gc := range giteaCommits {
		if gc.Sha == startHash {
			foundStart = true
		}
		if foundStart {
			commits = append(commits, Commit{
				Hash:    gc.Sha,
				Message: gc.Commit.Message,
				Author:  gc.Commit.Author.Name,
				Date:    gc.Commit.Author.Date,
			})
		}
		if gc.Sha == endHash {
			break
		}
	}
	fmt.Printf("Fetched %d commits from Gitea API\n", len(commits))

	return commits, nil
}

func (c *GiteaClient) GetCommitByHash(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) (*Commit, error) {
	var settings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse project settings: %w", err)
	}

	apiURL, ok := settings["git_api_url"].(string)
	if !ok {
		apiURL = c.baseURL
	}

	repo := project.RepositoryUrl
	url := fmt.Sprintf("%s/repos/%s/git/commits/%s", apiURL, repo, commitHash)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if project.AccessToken != "" {
		req.Header.Set("Authorization", "token "+project.AccessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("git API error: %s - %s", resp.Status, string(body))
	}

	var giteaCommit struct {
		Sha    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&giteaCommit); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Commit{
		Hash:    giteaCommit.Sha,
		Message: giteaCommit.Commit.Message,
		Author:  giteaCommit.Commit.Author.Name,
		Date:    giteaCommit.Commit.Author.Date,
	}, nil
}
