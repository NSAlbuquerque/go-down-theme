package vsplace

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"testing"
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
	if err != nil {
		t.Error(err)
	}

	t.Logf("%#v\n", data.Properties.toMap())
}

func TestParse(t *testing.T) {
	data, err := ioutil.ReadFile(path.Join(".", "testdata", "response1.json"))
	if err != nil {
		t.Error(err)
	}

	extensions, total, err := parseExtensions(bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
	}

	if len(extensions) == 0 {
		t.Fail()
	}

	if total < 0 {
		t.Fail()
	}

	t.Logf("Total de temas: %d\n", total)

	for i, ext := range extensions {
		t.Logf("%d - %s", i, ext.DisplayName)
	}
}

func TestProvider_fetchExtensions(t *testing.T) {
	p := NewProvider().(*provider)

	extensions, err := p.fetchExtensions()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Total de temas: %d\n", len(extensions))

	for i, ext := range extensions {
		t.Logf("%d - %s [%s]", i, ext.DisplayName, ext.Publisher.DisplayName)
	}
}
