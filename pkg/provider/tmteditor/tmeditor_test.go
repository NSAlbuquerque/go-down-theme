package tmteditor

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_GetGallery(t *testing.T) {
	p := NewProvider()

	gallery, err := p.GetGallery()
	if err != nil {
		t.Error(err)
	}

	for _, th := range gallery {
		t.Logf("%#v", th)
	}
}

func Test_parseEditorThemes(t *testing.T) {

	data, err := ioutil.ReadFile("./testdata/gallery.json")
	assert.NoError(t, err)

	themes, err := parseEditorThemes(bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(themes))
	assert.Less(t, 750, len(themes), "deve haver no minimo uns 750 temas na galeria")

	for _, th := range themes {
		assert.NotEmpty(t, th.Name, "o nome do tema é obrigatório")
		assert.NotEmpty(t, th.URL, "o URL de download do tema é obrigatório")
	}
}

func TestParseRepoFromURL(t *testing.T) {

	samples := []string{
		"https://raw.githubusercontent.com/idleberg/RetroComputers.tmTheme/master/Atari%20ST.tmTheme",
		"https://raw.githubusercontent.com/axar/Axar-SublimeTheme/master/Axar.tmTheme",
		"https://raw.githubusercontent.com/chriskempson/base16-textmate/master/themes/base16-ashes.light.tmTheme",
	}

	for _, s := range samples {
		repo, err := parseRepoFromURL(s)
		assert.NoError(t, err)
		assert.NotEmpty(t, repo.Owner, "o nome do dono do repositório é obrigatório")
	}

}
