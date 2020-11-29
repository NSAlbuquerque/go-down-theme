// Package pkgctrl implementa um provedor de temas para o Sublime Package Control (https://packagecontrol.io).
package pkgctrl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/albuquerq/go-down-theme/pkg/provider/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
)

const (
	providerName    = "Package Control"
	labelEndpoint   = "https://packagecontrol.io/browse/labels/"
	packageEndpoint = "https://packagecontrol.io/packages/"
)

// DefaultLabels labels com temas no Sublime Package Control.
var DefaultLabels = []string{"theme", "color scheme", "monokai"}

// DefaultRequestsInterval número de requisições por segundo.
const DefaultRequestsInterval = time.Second / 25

type provider struct {
	cli              *http.Client
	labels           []string
	requestsInterval time.Duration
	tiker            *time.Ticker
}

// NewProvider retorna um provedor de temas do Package Control.
// Busca somente por pacotes que estejam nos labels informados.
func NewProvider(labels []string) theme.Provider {
	return &provider{
		labels:           labels,
		cli:              http.DefaultClient,
		requestsInterval: DefaultRequestsInterval,
		tiker:            time.NewTicker(DefaultRequestsInterval),
	}
}

// NewProviderWithClient retorna um provedor de temas do Package Control com o
// cliente HTTP especificado. Busca somente pacotes que estejam nos labels informados.
func NewProviderWithClient(labels []string, cli *http.Client) theme.Provider {
	return &provider{
		labels: labels,
		cli:    cli,
	}
}

func (p *provider) SetRequestsInterval(interval time.Duration) {
	p.requestsInterval = interval
	p.tiker.Reset(p.requestsInterval)
}

// GetGallery retorna a galeria de temas.
// Em caso de erro retorna erro e a galeria de melhor esforço.
func (p *provider) GetGallery() (theme.Gallery, error) {
	names, err := p.fetchPackagesNames()
	if err != nil {
		return nil, err
	}

	pkgs, err := p.fetchPackages(names)
	if err != nil {
		return nil, err
	}

	var gallery theme.Gallery

	for _, pkg := range pkgs {

		if pkg.IsMissing {
			continue
		}

		var srcrepo string

		for _, src := range pkg.Sources {
			srcrepo = src
			repo, err := github.RepoFromURL(src)
			if err != nil {
				continue
			}
			if repo != nil {
				srcrepo = repo.String()
				break
			}
		}

		th := theme.Theme{
			Name:        pkg.Name,
			Description: pkg.Description,
			Author:      strings.Join(pkg.Authors, ", "),
			Provider:    providerName,
			Readme:      pkg.Readme,
			ProjectRepo: srcrepo,
			UpdatedAt:   pkg.LastModified,
		}

		if len(pkg.Versions) > 0 {
			th.Version = strconv.Itoa(pkg.Versions[len(pkg.Versions)-1])
		}

		gallery = append(gallery, th)
	}
	return gallery, nil
}

type pkg struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Homepage     string    `json:"homepage"`
	Authors      []string  `json:"authors"`
	Labels       []string  `json:"labels"`
	Versions     []int     `json:"st_versions"`
	LastModified time.Time `json:"last_modified"`
	IsMissing    bool      `json:"is_missing"`
	MissingError string    `json:"missing_error"`
	Sources      []string  `json:"sources"`
	Readme       string    `json:"readme"`
	Removed      bool      `json:"removed"`
	//Buy          interface{} `json:"buy"`
}

func parsePackage(r io.Reader) (pkg, error) {
	var pk pkg

	err := json.NewDecoder(r).Decode(&pk)
	if err != nil {
		return pkg{}, err
	}

	return pk, nil
}

func parsePackageNames(r io.Reader) ([]string, error) {
	pkgresp := struct {
		Packages []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}{}

	err := json.NewDecoder(r).Decode(&pkgresp)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(pkgresp.Packages))

	for _, pk := range pkgresp.Packages {
		names = append(names, pk.Name)
	}
	return names, nil
}

func (p *provider) fetchPackagesNames() ([]string, error) {
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	var (
		group errgroup.Group
		mux   sync.Mutex
	)

	pkgnames := []string{}

	for _, label := range p.labels {
		label := label

		group.Go(func() error {
			log.Println("Fetching label:", label)

			resp, err := p.cli.Get(labelEndpoint + label + ".json")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return errors.New("error on fetch packages for label" + label)
			}

			parsed, err := parsePackageNames(resp.Body)
			if err != nil {
				return err
			}

			mux.Lock()
			pkgnames = append(pkgnames, parsed...)
			mux.Unlock()

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return pkgnames, err
	}

	return pkgnames, nil
}

func (p *provider) fetchPackages(pkgnames []string) ([]pkg, error) {
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	var (
		mux   sync.Mutex
		group errgroup.Group
	)

	var pkgs []pkg

	for _, pkname := range pkgnames {
		pkname := pkname

		select {
		case <-p.tiker.C:
			group.Go(func() error {
				resp, err := p.cli.Get(packageEndpoint + pkname + ".json")
				if err != nil {
					return nil
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					if err != nil {
						log.Println("error on dump HTTP response", err)
						return err
					}

					// Pacote não existe mais.
					if resp.StatusCode == http.StatusNotFound {
						pk := pkg{
							Name:      pkname,
							IsMissing: true,
							Removed:   true,
						}
						mux.Lock()
						pkgs = append(pkgs, pk)
						mux.Unlock()
						return nil
					}

					return fmt.Errorf("erro on get the %s package", pkname)
				}

				if resp.Header.Get("content-type") != "application/json" {
					return fmt.Errorf("not json data returned for the %s package", pkname)
				}

				pk, err := parsePackage(resp.Body)
				if err != nil {
					return err
				}

				// Não adiciona pacotes faltosos.
				if pk.IsMissing {
					return nil
				}

				mux.Lock()
				pkgs = append(pkgs, pk)
				mux.Unlock()

				return nil
			})
		}
	}

	err := group.Wait()
	if err != nil {
		return pkgs, err
	}

	return pkgs, nil
}
