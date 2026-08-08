package viewmodel

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
	return `{{if .Stars}}<span id="gh-stars"><i class="ph ph-star"></i>{{.Stars}}</span>{{end}}`
}

func (vm *Stargazers) Load(req *http.Request) error {
	if vm.RepositoryURL == "" {
		return nil
	}

	svc, ok := torque.Inject[services.Services](req, ContextKeyServices)
	if !ok {
		return fmt.Errorf("failed to load dependency: services.Services")
	}

	count, err := svc.GitHubService.GetStars(req.Context(), vm.RepositoryURL)
	if err != nil {
		return err
	}
	vm.Stars = fmt.Sprintf("%d", count)

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
