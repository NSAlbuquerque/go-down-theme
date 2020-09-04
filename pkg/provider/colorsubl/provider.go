package colorsubl

import (
	"encoding/json"
	"net/http"

	"github.com/albuquerq/go-down-theme/pkg/theme"
)

const (
	galleryURL    = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes.json"
	themeFilePath = "https://raw.githubusercontent.com/Colorsublime/Colorsublime-Themes/master/themes/"
)

type csblprov struct{}

// NewProvider retorna um provedor de temas do ColorSublime.
func NewProvider() theme.Provider {
	return &csblprov{}
}

func (*csblprov) GetGallery() (gallery theme.Gallery, err error) {

	resp, err := http.Get(galleryURL)
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
			Provider: td.Author,
			URL:      themeFilePath + td.FileName,
		}

		gallery = append(gallery, t)
	}

	return
}
