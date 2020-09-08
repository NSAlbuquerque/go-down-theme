package pkgctrl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parsePackageNames(t *testing.T) {

	data, err := ioutil.ReadFile(path.Join(".", "testdata", "expackages.json"))
	if err != nil {
		t.Fatal(err)
	}

	pkgs, err := parsePackageNames(bytes.NewBuffer(data))
	assert.NoError(t, err)

	assert.Less(t, 0, len(pkgs))

	for _, pkg := range pkgs {
		t.Logf("%#v", pkg)
	}
}

func Test_parsePackages(t *testing.T) {

	data, err := ioutil.ReadFile(path.Join(".", "testdata", "expkg.json"))
	if err != nil {
		t.Fatal(err)
	}

	pkg, err := parsePackage(bytes.NewBuffer(data))
	assert.NoError(t, err)

	assert.NotEmpty(t, pkg.Name)
	t.Logf("%#v\n", pkg)
}

func TestProvider_fetchdata(t *testing.T) {
	var err error
	pknames := []string{}

	p := NewProvider(DefaultLabels).(*provider)

	t.Run("Fetch package names", func(t *testing.T) {
		pknames, err = p.fetchPackagesNames()
		assert.NoError(t, err)

		assert.Less(t, 0, len(pknames))

		for i, name := range pknames {
			assert.NotEmpty(t, name)
			t.Log(1+i, name)
		}
	})

	t.Run("Fetch packages from names", func(t *testing.T) {
		pkgs, err := p.fetchPackages(pknames)
		assert.NoError(t, err)

		t.Logf("Fetched %d packages at %d", len(pkgs), len(pknames))

		b := bytes.Buffer{}

		enc := json.NewEncoder(&b)
		enc.SetIndent("", "\t")

		err = enc.Encode(pkgs)
		assert.NoError(t, err)

		err = ioutil.WriteFile(path.Join(".", "testdata", "packages.json"), b.Bytes(), os.ModePerm)
		assert.NoError(t, err)
	})
}

func TestProvider_GetGalley(t *testing.T) {
	p := NewProvider(DefaultLabels)

	gallery, err := p.GetGallery()
	assert.NoError(t, err)

	assert.Less(t, 0, len(gallery))

	for i, th := range gallery {
		t.Logf("%d, - %#v", i+1, th)
	}
}
