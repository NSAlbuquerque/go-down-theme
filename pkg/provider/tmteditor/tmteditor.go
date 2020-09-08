package tmteditor

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/albuquerq/go-down-theme/pkg/provider/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
)

func init() {
	log.Println("not implemented")
}

const (
	providerName = "tmTheme-editor"
	sourceURL    = "https://tmtheme-editor.herokuapp.com/gallery.json"
)

type provider struct {
	cli *http.Client
}

// NewProvider retorna um provedor de temas para o tmTheme-editor.
func NewProvider() theme.Provider {
	return &provider{http.DefaultClient}
}

// NewProviderWithClient retorna um provedor de temas para o tmTheme-editor.
// Com um cliente HTTP específico.
func NewProviderWithClient(cli *http.Client) theme.Provider {
	return &provider{cli}
}

func (p *provider) GetGallery() (theme.Gallery, error) {
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}
	resp, err := p.cli.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("tm-theme-editor gallery not avaliable")
	}

	editorThemes, err := parseEditorThemes(resp.Body)

	total := len(editorThemes)

	if total == 0 {
		return nil, errors.New("themes not found")
	}

	gallery := make(theme.Gallery, 0, total)

	for _, tmtTheme := range editorThemes {

		th := theme.Theme{
			Name:     tmtTheme.Name,
			Provider: providerName,
			Author:   tmtTheme.Author,
			Light:    tmtTheme.Light,
			URL:      tmtTheme.URL,
		}

		repo, err := github.RepoFromURL(tmtTheme.URL)
		if err == nil {
			th.ProjectRepo = repo.String()
		} else {
			log.Println("fail on get repo info:", err)
		}

		gallery = append(gallery, th)
	}

	return gallery, nil
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
