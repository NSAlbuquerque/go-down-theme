package colorsubl

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/albuquerq/go-down-theme/pkg/theme"
)

const (
	galleryURL    = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes.json"
	themeFilePath = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes/"
	providerName  = "Color Sublime"
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
		Author   string
		FileName string
		Title    string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&themeData)
	if err != nil {
		return
	}

	gallery = make(theme.Gallery, 0, len(themeData))

	for _, td := range themeData {
		t := theme.Theme{
			Name:     td.Title,
			Author:   td.Author,
			Provider: providerName,
			URL:      themeFilePath + td.FileName,
		}

		gallery = append(gallery, t)
	}

	return
}
