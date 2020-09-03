package github

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestGithubProvider_findInternalThemeFiles(t *testing.T) {
	repo := Repo{
		Branch: "master",
		Owner:  "filmgirl",
		Name:   "TextMate-Themes",
	}

	p := NewProvider(&repo, "public_repo").(*provider)

	files, err := p.findInternalThemeFiles()
	if err != nil {
		t.Error(err)
	}
	assert.NoError(t, err)

	assert.Equal(t, 881, len(files))
}

func TestGithubProvider_GetGallery(t *testing.T) {
	repo := Repo{
		Branch: "master",
		Owner:  "kristopherjohnson",
		Name:   "MonochromeSublimeText",
	}

	p := NewProvider(&repo, "public_repo").(*provider)

	gallery, err := p.GetGallery()
	if err != nil {
		t.Error(err)
	}

	for _, th := range gallery {

		t.Logf("Baixando %s...", th.URL)

		b := bytes.Buffer{}
		err := common.Download(th.URL, &b)

		err = ioutil.WriteFile(path.Join("downloads", th.Name), b.Bytes(), os.ModePerm)
		if err != nil {
			t.Error(err)
		}
		t.Log("sucesso")
	}

}
