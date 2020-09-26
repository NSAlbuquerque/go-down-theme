package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	branchEndpointFmt = "https://api.github.com/repos/%s/%s/branches"
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

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			branchEndpointFmt,
			r.Owner,
			r.Name,
		),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(resp.Status, resp.StatusCode)
		return fmt.Errorf("error on fetch default branch")
	}

	branches := []struct {
		Name string `json:"name"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&branches)
	if err != nil {
		return err
	}

	for _, b := range branches {
		if b.Name == "master" || b.Name == "main" {
			r.Branch = b.Name
			return nil
		}
	}

	if r.Branch == "" {
		log.Println("default branch not found")
		return fmt.Errorf("default branch not found")
	}

	return nil
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
