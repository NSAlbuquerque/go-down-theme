package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	repoEndpointFmt = "https://api.github.com/repos/%s/%s"
)

// Github related errors.
var (
	ErrDefaultBranchNotFound = errors.New("default brach not found")
	ErrRateLimitExceeded     = errors.New("API hate limit has exceeded")
)

// Repo representa um repositório do github.
type Repo struct {
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Branch string `json:"branch"`
}

func (r Repo) String() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

// LoadDefaultBranch consulta o branch default a partir da API do github e o atribui ao repositório.
func (r *Repo) LoadDefaultBranch() error {
	if r.Branch != "" {
		return nil
	}

	repoData, err := r.fetchRepoData()
	if err != nil {
		return ErrDefaultBranchNotFound
	}

	if repoData.DefaultBranch == "" {
		return ErrDefaultBranchNotFound
	}

	r.Branch = repoData.DefaultBranch

	return nil
}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 403 {
			return ErrRateLimitExceeded
		}
		return ErrDefaultBranchNotFound
	}

type apiResponseType struct {
	License struct {
		Name string `json:"name"`
	} `json:"license"`

	DefaultBranch string `json:"default_branch"`
}

func (r *Repo) fetchRepoData() (*apiResponseType, error) {
	resp, err := http.Get(fmt.Sprintf(repoEndpointFmt, r.Owner, r.Name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 403 {
			return nil, ErrRateLimitExceeded
		}
		return nil, ErrDefaultBranchNotFound
	}
	var repoData apiResponseType
	err = json.NewDecoder(resp.Body).Decode(&repoData)
	if err != nil {
		return nil, err
	}
	return &repoData, nil
}

// InferReadme retorna o endereço mais provável do arquivo README.md do repositório.
// A existência do endereço não é verificada, ele pode ser um endereço inválido.
func (r Repo) InferReadme() string {
	err := r.LoadDefaultBranch()
	if err != nil {
		r.Branch = "master"
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md", r.Owner, r.Name, r.Branch)
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

	if ghurl.Path == "" {
		return nil, fmt.Errorf("invalid github URL %s", addr)
	}

	parts := strings.SplitN(ghurl.Path[1:], "/", 4)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid github URL: %s", addr)
	}

	repo := &Repo{
		Owner: parts[0],
		Name:  parts[1],
	}

	err = repo.LoadDefaultBranch()
	if err != nil {
		return repo, err
	}

	return repo, nil
}
