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
	ErrLicenseNotFound       = errors.New("license not found")
	ErrNoGithubRepository    = errors.New("it is not github repository")
)

// Repo representa um repositório do github.
type Repo struct {
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Branch  string `json:"branch"`
	License string `json:"license"`
}

func (r Repo) String() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

// LoadLicenseAndBranch consulta o branch default a partir da API do github e o atribui ao repositório.
func (r *Repo) LoadLicenseAndBranch() error {
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

	if r.License == "" {
		r.License = repoData.License.Name
	}
	r.Branch = repoData.DefaultBranch

	return nil
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
	err := r.LoadLicenseAndBranch()
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
		return nil, ErrNoGithubRepository
	}

	ghurl, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("inválid URL")
	}

	if ghurl.Path == "" {
		return nil, ErrNoGithubRepository
	}

	parts := strings.SplitN(ghurl.Path[1:], "/", 4)
	if len(parts) < 2 {
		return nil, ErrNoGithubRepository
	}

	repo := &Repo{
		Owner: parts[0],
		Name:  parts[1],
	}
	return repo, nil
}
