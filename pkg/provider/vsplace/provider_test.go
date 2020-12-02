package vsplace

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPropertiesParse(t *testing.T) {
	var jdata = `{
		"properties": [
        	{
				"key": "Microsoft.VisualStudio.Services.Links.Getstarted",
				"value": "https://github.com/uloco/theme-bluloco-light.git"
			},
			{
				"key": "Microsoft.VisualStudio.Services.Links.Source",
				"value": "https://github.com/uloco/theme-bluloco-light.git"
			},
			{
				"key": "Microsoft.VisualStudio.Services.Links.GitHub",
				"value": "https://github.com/uloco/theme-bluloco-light.git"
			},
			{
				"key": "Microsoft.VisualStudio.Code.Engine",
				"value": "^1.10.0"
			},
			{
				"key": "Microsoft.VisualStudio.Services.GitHubFlavoredMarkdown",
				"value": "true"
			},
			{
				"key": "Microsoft.VisualStudio.Code.ExtensionDependencies",
				"value": ""
			},
			{
				"key": "Microsoft.VisualStudio.Code.ExtensionPack",
				"value": ""
			},
			{
				"key": "Microsoft.VisualStudio.Code.LocalizedLanguages",
				"value": ""
			},
			{
				"key": "Microsoft.VisualStudio.Code.ExtensionKind",
				"value": "ui,workspace,web"
			}
		]}`

	data := struct {
		Properties properties `json:"properties"`
	}{
		Properties: properties{},
	}

	err := json.NewDecoder(strings.NewReader(jdata)).Decode(&data)
	assert.NoError(t, err)

	propMap := data.Properties.toMap()
	assert.NotNil(t, propMap)

	assert.Len(t, propMap, 9) // Número de propridades do JSON.

	for prop, value := range propMap {
		assert.NotEmpty(t, prop)
		t.Logf("%s => %v\n", prop, value)
	}
}

func TestParseExtensions(t *testing.T) {
	data, err := ioutil.ReadFile(path.Join(".", "testdata", "response1.json"))
	assert.NoError(t, err)

	p := NewProvider()

	extensions, total, err := p.parseExtensions(bytes.NewBuffer(data))
	assert.NoError(t, err)
	assert.Equal(t, 3593, total)

	assert.Len(t, extensions, 54)

	t.Logf("Total de temas: %d\n", total)

	for i, ext := range extensions {
		assert.NotEmpty(t, ext.Name)
		assert.NotEmpty(t, ext.DisplayName)
		t.Logf("%d - %s: %s", i, ext.Name, ext.DisplayName)
	}
}

func TestProvider_fetchExtensions(t *testing.T) {
	p := NewProvider()

	extensions, err := p.fetchExtensions()
	assert.NoError(t, err)

	assert.NotNil(t, extensions)
	assert.Greater(t, len(extensions), 0)

	t.Logf("Total de temas: %d\n", len(extensions))

	for i, ext := range extensions {
		assert.NotEmpty(t, ext.Name)
		t.Logf("%d - %s [%s]", i, ext.DisplayName, ext.Publisher.DisplayName)
	}
}

func TestProvider_GetGallery(t *testing.T) {
	p := NewProvider()
	p.SetRequestsInterval(500 * time.Millisecond)
	gallery, err := p.GetGallery()

	assert.NoError(t, err)

	assert.Greater(t, len(gallery), 0)

	for i, th := range gallery {
		t.Logf("%d - %s: %s | %s", i, th.Name, th.Description, th.ProjectRepoID)
	}

}
