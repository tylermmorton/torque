package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GitHubService interface {
	GetStars(ctx context.Context, repoURL string) (int, error)
}

func NewGitHubService() GitHubService {
	return &gitHubServiceImpl{client: &http.Client{}}
}

type gitHubServiceImpl struct {
	client *http.Client
}

func (s *gitHubServiceImpl) GetStars(ctx context.Context, repoURL string) (int, error) {
	owner, repo, err := parseRepoURL(repoURL)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var payload struct {
		StargazersCount int `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return payload.StargazersCount, nil
}

// parseRepoURL extracts owner and repo from a GitHub URL.
// Accepts https://github.com/owner/repo (with or without trailing slash/path).
func parseRepoURL(repoURL string) (owner, repo string, err error) {
	repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
	repoURL = strings.TrimPrefix(repoURL, "http://github.com/")
	parts := strings.SplitN(repoURL, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid github repository url: %s", repoURL)
	}
	return parts[0], parts[1], nil
}