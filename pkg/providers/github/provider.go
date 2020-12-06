package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/theme"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	providerName            = "Github"
	searchEndpointFmt       = "https://api.github.com/search/code?"
	defaultRequestsInterval = time.Minute / 10 // Intervalo entre as requisições.
)

type Provider struct {
	token  string
	repo   *Repo
	cli    *http.Client
	logger *logrus.Logger
	ticker *time.Ticker
}

// NewProvider returns a github theme provider.
func NewProvider(repo *Repo, opts ...Option) *Provider {
	p := &Provider{
		repo:   repo,
		cli:    http.DefaultClient,
		logger: logrus.New(),
		ticker: time.NewTicker(defaultRequestsInterval),
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Option applies options to the theme provider.
type Option func(p *Provider)

// WithToken applies the github auth token to the theme provider.
func WithToken(token string) Option {
	return func(p *Provider) {
		p.token = token
	}
}

// WithLogger applies custom logger to the theme provider.
func WithLogger(logger *logrus.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

// WithHTTPClient applies custom HTTP client to
// the theme provider.
func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

// SetRequestInterval sets the interval between requests.
func (p *Provider) SetRequestInterval(interval time.Duration) {
	p.ticker.Reset(interval)
}

// GetGallery return the theme gallery.
func (p *Provider) GetGallery() (theme.Gallery, error) {
	log := p.operation("Provider.GetGallery")
	files, err := p.findInternalThemeFiles()
	if err != nil {
		log.WithError(err).Println("themes not found")
		return nil, err
	}

	gallery := make(theme.Gallery, 0, len(files))

	for _, f := range files {
		t := theme.Theme{
			Name:          f.Name,
			Author:        p.repo.Owner,
			Provider:      providerName,
			ProjectRepoID: common.Hash(p.repo.String()),
			ProjectRepo:   p.repo.String(),
			URL:           f.DownloadURL,
		}

		gallery = append(gallery, t)
	}

	return gallery, nil
}

func (p *Provider) operation(operation string) *logrus.Entry {
	return p.logger.WithField("operation", operation)
}

func (p *Provider) findInternalThemeFiles() ([]File, error) {
	log := p.operation("Provider.findInternalThemeFiles")
	var (
		pg    = 1
		perPg = 100
		pgs   = 0
	)

	total, files, err := p.fetch(pg, perPg)
	if err != nil || total == 0 {
		log.WithField("total", total).WithError(err).Println("request error")
		return nil, err
	}

	if total > perPg {
		pgs = int(math.Ceil(float64(total) / float64(perPg)))
	}

	group := errgroup.Group{}
	mux := sync.Mutex{}

	for pg = 2; pg <= pgs; pg++ {
		page := pg
		select {
		case <-p.ticker.C:
			group.Go(func() error {
				_, pageFiles, err := p.fetch(page, perPg)
				if err != nil {
					p.logger.WithError(err).Printf("erro on fetch page %d", page)
					return err
				}
				mux.Lock()
				files = append(files, pageFiles...)
				mux.Unlock()
				return nil
			})
		}
	}

	err = group.Wait()
	if err != nil {
		log.WithError(err).Println("error in some request")
		return nil, err
	}
	return files, nil
}

func (p *Provider) fetch(page, perPage int) (total int, files []File, err error) {
	log := p.operation("Provider.fetch")
	if p.cli == nil {
		err = errors.New("the http client must be specified")
		return
	}

	query := make(url.Values)
	query.Set("q", fmt.Sprintf("repo:%s/%s extension:tmTheme", p.repo.Owner, p.repo.Name))
	query.Set("page", strconv.Itoa(page))
	query.Set("per_page", strconv.Itoa(perPage))

	endpoint := searchEndpointFmt + query.Encode()

	log.Printf("processing %s URL", endpoint)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.WithError(err).Error("erro on prepare request")
		return
	}

	req.Header.Set("accept", "application/vnd.github.v3+json")
	if p.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", p.token))
	}

	resp, err := p.cli.Do(req)
	if err != nil {
		log.WithError(err).WithField("code", resp.StatusCode).Println("request error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("statusCode", resp.StatusCode).WithField("url", endpoint).Error("request error")
		err = errors.New("error on perform search request")
		return
	}

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
		log.WithError(err).Error("error on decode response body")
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
