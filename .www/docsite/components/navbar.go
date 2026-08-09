package components

import (
	"fmt"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/services"
)

type Stargazers struct {
	RepositoryURL string
	Stars         string
}

func (*Stargazers) StyleSheet() string {
	//language=css
	return `
#gh-stars {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-left: 6px;
  border-left: 1px solid var(--color-neutral-700);
  margin-left: 2px;
}
#gh-stars .ph-star {
  color: var(--color-accent);
}
`
}

func (*Stargazers) Template() string {
	//language=html
	return `{{if .Stars}}<span id="gh-stars" data-test-id="stars-container"><i class="ph ph-star"></i><span data-test-id="star-count">{{.Stars}}</span></span>{{end}}`
}

func (s *Stargazers) Load(req *http.Request) error {
	if s.RepositoryURL == "" {
		return nil
	}

	svc, err := torque.Inject[*services.Services](req, ContextKeyServices)
	if err != nil {
		return fmt.Errorf("failed to load dependency: %w", err)
	}

	count, err := svc.GitHubService.GetStars(req.Context(), s.RepositoryURL)
	if err != nil {
		return err
	}
	s.Stars = fmt.Sprintf("%d", count)

	return nil
}

type NavBar struct {
	GitHubURL string
	Stars     Stargazers `template:"github-stars"`
}

func (n *NavBar) Load(req *http.Request) error {
	n.Stars.RepositoryURL = n.GitHubURL
	return nil
}
