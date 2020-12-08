package tmteditor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
	"github.com/albuquerq/go-down-theme/pkg/providers/github"
)

const (
	// Name of provider.
	Name themes.ProviderName = "tmTheme-editor"

	sourceURL = "https://tmtheme-editor.herokuapp.com/gallery.json"
)

// Provider tmTheme-editor provider.
type Provider struct {
	cli    *http.Client
	logger *logrus.Logger
}

var _ themes.Provider = &Provider{}

// NewProvider returns a theme provider for tmTheme-editor.
func NewProvider(opts ...Option) *Provider {
	p := &Provider{
		cli:    http.DefaultClient,
		logger: logrus.New(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Option apply option to provider.
type Option func(p *Provider)

// WithLogger returns a custom logger option.
func WithLogger(logger *logrus.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

// WithHTTPClient returns a custom HTTP client option.
func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

// GetGallery returns the tmTheme-editor theme gallery.
func (p *Provider) GetGallery() (themes.Gallery, error) {
	log := p.operation("Provider.GetGallery")

	resp, err := p.cli.Get(sourceURL)
	if err != nil {
		log.WithError(err).Println("error on fetch gallery resource")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("statusCode", resp.StatusCode).Println("response status error")
		return nil, errors.New("tm-theme-editor gallery not avaliable")
	}

	editorThemes, err := parseEditorThemes(resp.Body)
	if err != nil {
		log.WithError(err).Println("error on proccess response data")
		return nil, err
	}

	total := len(editorThemes)

	if total == 0 {
		log.Println("themes not found")
		return nil, errors.New("themes not found")
	}

	gallery := make(themes.Gallery, 0, total)

	for _, tmtTheme := range editorThemes {

		th := themes.Theme{
			Name:     tmtTheme.Name,
			Provider: Name,
			Author:   tmtTheme.Author,
			Light:    tmtTheme.Light,
			URL:      tmtTheme.URL,
		}

		repo, err := github.RepoFromURL(tmtTheme.URL)
		if err == nil {
			th.ProjectRepo = repo.String()
			th.ProjectRepoID = common.Hash(th.ProjectRepo)
		} else {
			log.Println("fail on get repo info:", err)
		}

		gallery = append(gallery, th)
	}

	return gallery, nil
}

func (p *Provider) operation(operation string) *logrus.Entry {
	return p.logger.WithField("operation", operation)
}

type editorTheme struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Light  bool   `json:"light"`
	Author string `json:"author"`
}

func parseEditorThemes(dataSource io.Reader) ([]editorTheme, error) {
	editorThemes := []editorTheme{}

	err := json.NewDecoder(dataSource).Decode(&editorThemes)
	if err != nil {
		return nil, err
	}
	return editorThemes, nil
}
