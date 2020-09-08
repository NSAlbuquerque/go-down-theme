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

func TestRepoFromURL(t *testing.T) {

	samples := []string{
		"https://raw.githubusercontent.com/idleberg/RetroComputers.tmTheme/master/Atari%20ST.tmTheme",
		"https://raw.githubusercontent.com/axar/Axar-SublimeTheme/master/Axar.tmTheme",
		"https://raw.githubusercontent.com/chriskempson/base16-textmate/master/themes/base16-ashes.light.tmTheme",
	}

	for _, s := range samples {
		repo, err := RepoFromURL(s)
		assert.NoError(t, err)
		assert.NotEmpty(t, repo.Owner, "o nome do dono do repositório é obrigatório")
		t.Log(repo)
	}

}
