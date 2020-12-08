package colorsubl

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
	"github.com/albuquerq/go-down-theme/pkg/providers/github"
)

const (
	// Name of provider.
	Name themes.ProviderName = "Color Sublime"

	galleryURL    = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes.json"
	themeFilePath = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes/"
)

var _ themes.Provider = &Provider{}

// Provider for color sublime project.
type Provider struct {
	cli    *http.Client
	logger *logrus.Logger
}

// NewProvider returns a ColorSublime theme provider.
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

// Option applies options to the provider.
type Option func(*Provider)

// WithLogger applies custom logger to the theme provider.
func WithLogger(logger *logrus.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

// WithHTTPClient applies custom HTTP client to the theme provider.
func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

// GetGallery returns the gallery of themes of color sublime project.
func (p *Provider) GetGallery() (gallery themes.Gallery, err error) {
	log := p.logger.WithField("operation", "Provider.GetGallery")
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	resp, err := p.cli.Get(galleryURL)
	if err != nil {
		log.WithError(err).Error("error on fetch http request")
		return
	}
	defer resp.Body.Close()

	themeData := []struct {
		Author      string
		Description string
		FileName    string
		Title       string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&themeData)
	if err != nil {
		log.WithError(err).Error("error on parse response body data")
		return
	}

	gallery = make(themes.Gallery, 0, len(themeData))

	repo, err := github.RepoFromURL(galleryURL)
	if err != nil {
		log.WithError(err).Error("error on parse github repository")
		return nil, err
	}

	for _, td := range themeData {

		t := themes.Theme{
			Name:          td.Title,
			Author:        td.Author,
			Description:   td.Description,
			Provider:      Name,
			URL:           themeFilePath + td.FileName,
			ProjectRepoID: common.Hash(repo.String()),
			ProjectRepo:   repo.String(),
		}

		gallery = append(gallery, t)
	}

	return
}
