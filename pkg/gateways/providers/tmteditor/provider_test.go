package tmteditor

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_GetGallery(t *testing.T) {
	p := NewProvider()

	assert.NotNil(t, p)
	assert.NotNil(t, p.cli)
	assert.NotNil(t, p.logger)

	gallery, err := p.GetGallery()
	assert.NoError(t, err)

	assert.NotNil(t, gallery)
	assert.Greater(t, len(gallery), 0)

	t.Logf("found %d themes", len(gallery))

	for _, th := range gallery {
		t.Logf("%#v", th)
	}
}

func Test_parseEditorThemes(t *testing.T) {

	data, err := ioutil.ReadFile("./testdata/gallery.json")
	assert.NoError(t, err)

	themes, err := parseEditorThemes(bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.Equal(t, len(themes), 768)

	log.Printf("parsed %d themes", len(themes))
	for _, th := range themes {
		assert.NotEmpty(t, th.Name, "the theme name is required")
		assert.NotEmpty(t, th.URL, "the theme URL is required")
	}
}
