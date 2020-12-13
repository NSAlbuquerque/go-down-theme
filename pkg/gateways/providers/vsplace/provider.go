package vsplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/domain/providers"
	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
	"github.com/albuquerq/go-down-theme/pkg/domain/vos"
	"github.com/sirupsen/logrus"
)

const (
	// Name of provider.
	Name vos.ProviderName = "Visual Studio Marketplace"

	extensionsEndpoint = "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery"

	pgSize                  = 100
	defaultRequestsInterval = time.Second / 5
)

// Provider represents a theme provider
// to the Visual Studio Marketplace.
type Provider struct {
	cli    *http.Client
	ticker *time.Ticker
	logger *logrus.Entry
}

var _ providers.Provider = &Provider{}

// Option applies option to the provider.
type Option func(*Provider)

// NewProvider return a theme provider to the Visual Studio Marketplace.
func NewProvider(options ...Option) *Provider {

	p := &Provider{
		ticker: time.NewTicker(defaultRequestsInterval),
		logger: logrus.NewEntry(logrus.New()),
		cli:    http.DefaultClient,
	}

	for _, opt := range options {
		opt(p)
	}
	return p
}

// WithLogger applies a custom logger to the provider.
func WithLogger(logger *logrus.Entry) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

// WithHTTPClient applies a custom HTTP client to the
// theme provider.
func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

// GetGallery returns the theme gallery.
func (p *Provider) GetGallery() (themes.Gallery, error) {
	log := p.operation("Provider.GetGallery")

	exts, err := p.fetchExtensions()
	if err != nil {
		log.WithError(err).Println("error on fetch extensions")
		return nil, err
	}

	var gallery themes.Gallery

	log.Printf("feteched %d extensions", len(exts))

	for _, ext := range exts {
		version := ext.Versions[0]

		repo := version.Properties.Get("Microsoft.VisualStudio.Services.Links.Source")

		t := themes.Theme{
			Author:        ext.Publisher.Name,
			Provider:      Name,
			Version:       version.Version,
			Name:          ext.DisplayName,
			Description:   ext.Description,
			ProjectRepoID: common.Hash(repo),
			ProjectRepo:   repo,
			Readme:        version.Properties.Get("Microsoft.VisualStudio.Services.Links.Learn"),
			LastUpdate:    version.LastUpdated,
		}

		branding := strings.ToLower(version.Properties.Get("Microsoft.VisualStudio.Services.Branding.Theme"))
		t.Light = branding == "light"

		gallery = append(gallery, t)
	}

	return gallery, nil
}

// SetRequestsInterval sets the interval between requests for
// the Visual Studio Marketplace API.
func (p *Provider) SetRequestsInterval(interval time.Duration) {
	p.ticker.Reset(interval)
}

type extension struct {
	Publisher struct {
		Name        string `json:"publisherName"`
		DisplayName string `json:"displayName"`
	}

	ID            string    `json:"extensionId"`
	Name          string    `json:"extensionName"`
	DisplayName   string    `json:"displayName"`
	LastUpdated   time.Time `json:"lastUpdated"`
	PublishedDate time.Time `json:"publishedDate"`
	Description   string    `json:"shortDescription"`

	Versions []version `json:"versions"`
}

type version struct {
	Version     string    `json:"version"`
	LastUpdated time.Time `json:"lastUpdated"`
	Files       []struct {
		AssetType string `json:"assetType"`
		Source    string `json:"source"`
	}
	Properties properties `json:"properties"` // Deve ser instanciado antes de decodificar JSON.
}

type properties []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (prs properties) toMap() map[string]string {
	m := make(map[string]string)

	for _, p := range prs {
		m[p.Key] = p.Value
	}
	return m
}

func (prs properties) Get(key string) string {

	for _, p := range prs {
		if key == p.Key {
			return p.Value
		}
	}

	return ""
}

const requestBodyFmt = `{
  "assetTypes": [
    "Microsoft.VisualStudio.Services.Icons.Default",
    "Microsoft.VisualStudio.Services.Icons.Branding",
    "Microsoft.VisualStudio.Services.Icons.Small"
  ],
  "filters": [
    {
      "criteria": [
        {
          "filterType": 8,
          "value": "Microsoft.VisualStudio.Code"
        },
        {
          "filterType": 10,
          "value": "target:\"Microsoft.VisualStudio.Code\" "
        },
        {
          "filterType": 12,
          "value": "37888"
        },
        {
          "filterType": 5,
          "value": "Themes"
        }
      ],
      "direction": 2,
      "pageSize": %d,
      "pageNumber": %d,
      "sortBy": 4,
      "sortOrder": 0,
      "pagingToken": null
    }
  ],
  "flags": 870
}`

func (p *Provider) fetchExtensions() ([]extension, error) {
	log := p.operation("Provider.fetchExtensions")

	extensions, total, err := p.fetchPageExtensions(1, pgSize)
	if err != nil {
		log.WithError(err).Println("error on fetch extensions page")
		return nil, err
	}

	if total == 0 {
		log.Warning("extensions not found")
		return nil, nil
	}

	pages := int(math.Ceil(float64(total) / float64(pgSize)))
	log.Printf("found %d extensions pages", pages)

	var (
		group errgroup.Group
		mux   sync.Mutex
	)

	for pg := 2; pg <= pages; pg++ {
		page := pg

		select {
		case <-p.ticker.C:
			group.Go(func() error {
				log.Printf("fetching extension page %d", page)
				exts, _, err := p.fetchPageExtensions(page, pgSize)
				if err != nil {
					log.WithError(err).Printf("error on fetch page %d", page)
					return err
				}

				mux.Lock()
				extensions = append(extensions, exts...)
				mux.Unlock()

				return nil
			})
		}
	}

	log.Println("all pages were fetched, waiting responses")
	err = group.Wait()
	if err != nil {
		log.WithError(err).Println("error in some searched page")
		return extensions, err
	}
	log.Println("all extensions were fetched")

	return extensions, nil
}

func (p *Provider) operation(op string) *logrus.Entry {
	return p.logger.WithField("operation", op)
}

func (p *Provider) fetchPageExtensions(pg, pgSize int) ([]extension, int, error) {
	log := p.operation("Provider.fetchPageExtensions")

	if p.cli == nil {
		return nil, -1, errors.New("http client not found")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		extensionsEndpoint,
		strings.NewReader(fmt.Sprintf(requestBodyFmt, pgSize, pg)),
	)
	if err != nil {
		return nil, -1, err
	}
	defer req.Body.Close()

	req.Header.Set("Accept", "application/json;api-version=6.1-preview.1;excludeUrls=true")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.cli.Do(req)
	if err != nil {
		log.WithError(err).Println("erro on fetch extensions")
		return nil, -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := errors.New("error on fetch extensions")
		log.WithError(err).Printf("error on get extensions; status code %d", resp.StatusCode)
		return nil, -1, err
	}

	return p.parseExtensions(resp.Body)
}

func (p *Provider) parseExtensions(in io.Reader) ([]extension, int, error) {
	data := struct {
		Results []struct {
			Extensions []extension `json:"extensions"`
			MetaData   []struct {
				Type  string `json:"metadataType"`
				Items []struct {
					Name  string `json:"name"`
					Count int    `json:"count"`
				} `json:"metadataItems"`
			} `json:"resultMetadata"`
		} `json:"results"`
	}{}

	err := json.NewDecoder(in).Decode(&data)
	if err != nil {
		return nil, -1, err
	}

	if len(data.Results) == 0 {
		return nil, -1, errors.New("no result for extensions")
	}

	total := -1

	if len(data.Results[0].MetaData) == 0 {
		return nil, -1, errors.New("not found metadata")
	}

	for _, i := range data.Results[0].MetaData[0].Items {
		if i.Name == "TotalCount" {
			total = i.Count
			break
		}
	}

	if total < 0 {
		return nil, -1, errors.New("not found total")
	}

	return data.Results[0].Extensions, total, err
}
