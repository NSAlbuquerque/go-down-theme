package github

import (
	"fmt"
	"net/url"
	"strings"
)

// Repo represents a theme repository.
type Repo struct {
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Branch string `json:"branch"`
}

func (r Repo) String() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

// InferReadme retorna o endereço mais provável do arquivo README.md do repositório.
// A existência do endereço não é verificada, ele pode ser um endereço inválido.
func (r Repo) InferReadme() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/README.md", r.Owner, r.Name)
}

// File ...
type File struct {
	Name        string
	DownloadURL string
}

// RepoFromURL retorna um repositório a partir
// de um URL do github.
func RepoFromURL(addr string) (*Repo, error) {
	if !strings.Contains(addr, "github.com") &&
		!strings.Contains(addr, "githubusercontent.com") {
		return nil, fmt.Errorf("invalid github URL: %s", addr)
	}

	ghurl, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid github URL %s: %v", addr, err)
	}

	parts := strings.SplitN(ghurl.Path[1:], "/", 4)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid github URL: %s", addr)
	}

	repo := &Repo{
		Owner:  parts[0],
		Name:   parts[1],
		Branch: parts[2],
	}

	return repo, nil
}
