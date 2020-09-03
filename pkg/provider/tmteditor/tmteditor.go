package tmteditor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/albuquerq/go-down-theme/pkg/provider/github"

	"github.com/albuquerq/go-down-theme/pkg/theme"
)

func init() {
	log.Println("not implemented")
}

const (
	sourceURL    = "https://tmtheme-editor.herokuapp.com/gallery.json"
	providerName = "tmTheme-editor"
)

type provider struct{}

// NewProvider returns a tmTheme-editor theme provider.
func NewProvider() theme.Provider {
	return &provider{}
}

func (*provider) GetGallery() (theme.Gallery, error) {
	resp, err := http.Get(sourceURL)
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
			Name:  tmtTheme.Name,
			Light: tmtTheme.Light,
			URL:   tmtTheme.URL,
		}

		repo, err := parseRepoFromURL(tmtTheme.URL)
		if err == nil {
			th.ProjectRepo = repo.String()
			th.Provider = repo.Owner
		} else {
			log.Println("fail on get repo info:", err)
		}

		gallery = append(gallery, th)
	}

	return gallery, nil
}

func parseRepoFromURL(sourceURL string) (*github.Repo, error) {
	purl, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}

	if purl.Hostname() != "raw.githubusercontent.com" {
		return nil, errors.New("invalid github raw file url: " + sourceURL)
	}

	parts := strings.SplitN(purl.Path[1:], "/", 4)

	if len(parts) != 4 {
		return nil, errors.New("invalid github raw file url")
	}

	repoURL := url.URL{
		Scheme: purl.Scheme,
		Host:   "github.com",
	}

	repo := &github.Repo{
		Owner:  parts[0],
		Name:   parts[1],
		Branch: parts[2],
	}

	repoURL.Path = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	return repo, nil
}

type editorTheme struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Light      bool   `json:"light"`
	Author     string `json:"author"`
	Maintainer string `json:"maintainer"`
}

func parseEditorThemes(dataSource io.Reader) ([]editorTheme, error) {
	editorThemes := []editorTheme{}

	err := json.NewDecoder(dataSource).Decode(&editorThemes)
	if err != nil {
		return nil, err
	}
	return editorThemes, nil
}
