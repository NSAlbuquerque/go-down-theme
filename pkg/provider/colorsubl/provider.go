package colorsubl

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/provider/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
	"github.com/sirupsen/logrus"
)

const (
	providerName  = "Color Sublime"
	galleryURL    = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes.json"
	themeFilePath = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes/"
)

type Provider struct {
	cli    *http.Client
	logger *logrus.Logger
}

// NewProvider retorna um provedor de temas do ColorSublime.
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

type Option func(*Provider)

func WithLogger(logger *logrus.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

func (p *Provider) GetGallery() (gallery theme.Gallery, err error) {
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

	gallery = make(theme.Gallery, 0, len(themeData))

	repo, err := github.RepoFromURL(galleryURL)
	if err != nil {
		log.WithError(err).Error("error on parse github repository")
		return nil, err
	}

	for _, td := range themeData {

		t := theme.Theme{
			Name:          td.Title,
			Author:        td.Author,
			Description:   td.Description,
			Provider:      providerName,
			URL:           themeFilePath + td.FileName,
			ProjectRepoID: common.Hash(repo.String()),
			ProjectRepo:   repo.String(),
		}

		gallery = append(gallery, t)
	}

	return
}
