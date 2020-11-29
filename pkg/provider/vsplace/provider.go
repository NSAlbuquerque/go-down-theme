package vsplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/albuquerq/go-down-theme/pkg/provider/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
)

const (
	providerName       = "Visual Studio Marketplace"
	extensionsEndpoint = "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery"

	pgSize     = 100
	reqAverage = 5
)

type provider struct {
	cli *http.Client
}

// NewProvider retorna um provedor de temas para
func NewProvider() theme.Provider {
	return &provider{
		cli: http.DefaultClient,
	}
}

func (p *provider) GetGallery() (theme.Gallery, error) {
	exts, err := p.fetchExtensions()
	if err != nil {
		return nil, err
	}

	var gallery theme.Gallery

	for _, ext := range exts {
		version := ext.Versions[0]

		t := theme.Theme{
			Author:      ext.Publisher.Name,
			Provider:    providerName,
			Version:     version.Version,
			Name:        ext.DisplayName,
			Description: ext.Description,
			ProjectRepo: version.Properties.Get("Microsoft.VisualStudio.Services.Links.Source"),
			Readme:      version.Properties.Get("Microsoft.VisualStudio.Services.Links.Learn"),
			LastUpdate:  version.LastUpdated,
		}

		branding := strings.ToLower(version.Properties.Get("Microsoft.VisualStudio.Services.Branding.Theme"))
		t.Light = branding == "light"

		if strings.Contains(t.ProjectRepo, "github") {
			repo, err := github.RepoFromURL(t.ProjectRepo)
			if err == nil {
				t.Readme = repo.InferReadme()
			}
		}

		gallery = append(gallery, t)
	}
	return gallery, nil
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

func (p *provider) fetchExtensions() ([]extension, error) {

	extensions, total, err := p.fetchPageExtensions(1, pgSize)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	pages := int(math.Ceil(float64(total) / float64(pgSize)))

	var (
		group errgroup.Group
		mux   sync.Mutex
	)

	for pg := 2; pg <= pages; pg++ {
		pg := pg
		group.Go(func() error {
			exts, _, err := p.fetchPageExtensions(pg, pgSize)
			if err != nil {
				return err
			}

			mux.Lock()
			extensions = append(extensions, exts...)
			mux.Unlock()

			return nil
		})
		time.Sleep(time.Second / reqAverage)
	}

	err = group.Wait()
	if err != nil {
		return extensions, err
	}

	return extensions, nil
}

func (p *provider) fetchPageExtensions(pg, pgSize int) ([]extension, int, error) {
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
		log.Println("erro on fetch extensions", err)
		return nil, -1, err
	}

	if resp.StatusCode != http.StatusOK {
		data, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(data))
		log.Printf("error on get extensions, code %d", resp.StatusCode)
		return nil, -1, errors.New("error on fetch extensions")
	}

	return parseExtensions(resp.Body)
}

func parseExtensions(in io.Reader) ([]extension, int, error) {

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
