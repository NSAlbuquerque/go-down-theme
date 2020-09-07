package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/albuquerq/go-down-theme/pkg/theme"
	"golang.org/x/sync/errgroup"
)

const (
	searchEndpointFmt = "https://api.github.com/search/code?q=+extension:.tmTheme+repo:%s/%s&page=%d&per_page=%d"
)

type provider struct {
	token string
	repo  *Repo
	cli   *http.Client
}

// NewProvider retorna um provedor de temas para um repositório do github.
func NewProvider(repo *Repo, token string) theme.Provider {
	return &provider{
		repo: repo,
		cli:  http.DefaultClient,
	}
}

// NewProviderWithClient retorna um provedor de temas para um repositório do github.
// Com um cliente HTTP específico.
func NewProviderWithClient(repo *Repo, token string, cli *http.Client) theme.Provider {
	return &provider{
		repo: repo,
		cli:  cli,
	}
}

func (p *provider) GetGallery() (theme.Gallery, error) {

	files, err := p.findInternalThemeFiles()
	if err != nil {
		return nil, err
	}

	gallery := make(theme.Gallery, 0, len(files))

	for _, f := range files {
		t := theme.Theme{
			Name:        f.Name,
			ProjectRepo: p.repo.String(),
			Provider:    p.repo.Owner,
			URL:         f.DownloadURL,
		}

		gallery = append(gallery, t)
	}

	return gallery, nil
}

func (p *provider) findInternalThemeFiles() ([]File, error) {

	// TODO: caso não tenha branch definido, fazer consulta na API.

	var (
		page    = 1
		perPage = 100
		pages   = 0
	)

	total, files, err := p.fetch(page, perPage)
	if err != nil || total == 0 {
		return nil, err
	}

	if total > perPage {
		pages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	group := errgroup.Group{}
	mux := sync.Mutex{}

	for page = 2; page <= pages; page++ {

		page := page

		group.Go(func() error {
			_, pageFiles, err := p.fetch(page, perPage)
			if err != nil {
				return err
			}
			mux.Lock()
			files = append(files, pageFiles...)
			mux.Unlock()
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (p *provider) fetch(page, perPage int) (total int, files []File, err error) {
	if p.cli == nil {
		err = errors.New("the http client must be specified")
		return
	}
	var endpoint = fmt.Sprintf(
		searchEndpointFmt,
		p.repo.Owner,
		p.repo.Name,
		page,
		perPage,
	)

	// log.Println("Fetch page:", page)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return
	}

	if p.token != "" {
		req.Header.Add("Authorization", "token "+p.token)
	}

	resp, err := p.cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("error on perform search request")
		return
	}

	// log.Printf(
	// 	"Rate limit: %s\nRemaining: %s\nReset: %s\nLinks: %s\n",
	// 	resp.Header.Get("X-RateLimit-Limit"),
	// 	resp.Header.Get("X-RateLimit-Remaining"),
	// 	resp.Header.Get("X-RateLimit-Reset"),
	// 	resp.Header.Get("Link"),
	// )

	searchResult := struct {
		Total int `json:"total_count"`
		Items []struct {
			FileName    string `json:"name"`
			Path        string `json:"path"`
			BranchesURL string `json:"branches_url"`
		} `json:"items"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return
	}

	total = searchResult.Total

	if total == 0 {
		return
	}

	files = make([]File, 0, len(searchResult.Items))

	for _, item := range searchResult.Items {
		f := File{
			Name: item.FileName,
			DownloadURL: strings.Join([]string{
				"https://raw.githubusercontent.com",
				p.repo.Owner,
				p.repo.Name,
				p.repo.Branch,
				item.Path,
			},
				"/",
			),
		}
		files = append(files, f)
	}
	return
}
