package colorsubl

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/provider/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
)

const (
	providerName  = "Color Sublime"
	galleryURL    = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes.json"
	themeFilePath = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes/"
)

type provider struct {
	cli *http.Client
}

// NewProvider retorna um provedor de temas do ColorSublime.
func NewProvider() theme.Provider {
	return &provider{
		cli: http.DefaultClient,
	}
}

// NewProviderWithClient retorna um provedor de temas do ColorSublime com um cliente HTTP específico.
func NewProviderWithClient(cli *http.Client) theme.Provider {
	return &provider{
		cli: cli,
	}
}

func (p *provider) GetGallery() (gallery theme.Gallery, err error) {
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	resp, err := p.cli.Get(galleryURL)
	if err != nil {
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
		return
	}

	gallery = make(theme.Gallery, 0, len(themeData))

	repo, err := github.RepoFromURL(galleryURL)
	if err != nil {
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
			Readme:        repo.InferReadme(),
		}

		gallery = append(gallery, t)
	}

	return
}
