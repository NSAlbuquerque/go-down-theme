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

func init() {
	accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
}

var accessToken string

func TestGithubProvider_findInternalThemeFiles(t *testing.T) {
	repo := Repo{
		Branch: "master",
		Owner:  "filmgirl",
		Name:   "TextMate-Themes",
	}

	p := NewProvider(&repo, WithToken(accessToken))

	files, err := p.findInternalThemeFiles()
	assert.NoError(t, err)

	assert.Equal(t, 155, len(files))
}

func TestGithubProvider_GetGallery(t *testing.T) {

	err := os.MkdirAll(path.Join(".", "downloads"), os.ModePerm)
	assert.NoError(t, err)

	repo := Repo{
		Branch: "master",
		Owner:  "kristopherjohnson",
		Name:   "MonochromeSublimeText",
	}

	p := NewProvider(&repo, WithToken(accessToken))

	gallery, err := p.GetGallery()
	assert.NoError(t, err)
	assert.Greater(t, len(gallery), 0)

	for _, th := range gallery {

		t.Logf("Baixando %s...", th.URL)

		b := bytes.Buffer{}
		err := common.Download(th.URL, &b)

		err = ioutil.WriteFile(path.Join(".", "downloads", th.Name), b.Bytes(), os.ModeSetuid|os.ModePerm)
		assert.NoError(t, err)
	}

}

func TestRepoFromURL(t *testing.T) {

	cases := []struct {
		in  string
		err bool
	}{
		{"https://packagecontrol.io/repository.json", true},
		{"https://github.com/idleberg/WarpOS.tmTheme", false},
		{"https://packagecontrol.io/repository.json", true},
		{"https://github.com/ctf0/Yeti_ST3", false},
		{"https://packagecontrol.io/repository.json", true},
		{"https://github.com/anton-rudeshko/sublime-yandex-wiki", false},
		{"https://packagecontrol.io/repository.json", true},
		{"https://github.com/jrvieira/zero-dark", false},
		{"https://packagecontrol.io/repository.json", true},
		{"https://raw.githubusercontent.com/idleberg/RetroComputers.tmTheme/master/Atari%20ST.tmTheme", false},
		{"https://raw.githubusercontent.com/axar/Axar-SublimeTheme/master/Axar.tmTheme", false},
		{"https://raw.githubusercontent.com/chriskempson/base16-textmate/master/themes/base16-ashes.light.tmTheme", false},
	}

	for _, c := range cases {
		repo, err := RepoFromURL(c.in)
		assert.Equal(t, c.err, err != nil)

		if c.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotEmpty(t, repo.Owner, "o nome do dono do repositório é obrigatório")
			t.Log(repo, repo.Branch)
		}
	}
}
