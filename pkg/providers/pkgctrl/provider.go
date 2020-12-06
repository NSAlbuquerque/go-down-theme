// Package pkgctrl implements a theme provider for Sublime Package Control (https://packagecontrol.io).
package pkgctrl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/albuquerq/go-down-theme/pkg/common"
	"github.com/albuquerq/go-down-theme/pkg/providers/github"
	"github.com/albuquerq/go-down-theme/pkg/theme"
	"github.com/sirupsen/logrus"
)

const (
	providerName    = "Package Control"
	labelEndpoint   = "https://packagecontrol.io/browse/labels/"
	packageEndpoint = "https://packagecontrol.io/packages/"
)

// DefaultLabels labels for themes in Sublime Package Control.
var DefaultLabels = []string{"theme", "color scheme", "monokai"}

// defaultRequestsInterval interval between HTTP requests.
const defaultRequestsInterval = time.Second / 25

// Provider para Sublime Package Control.
type Provider struct {
	labels []string
	cli    *http.Client
	tiker  *time.Ticker
	logger *logrus.Logger
}

var _ theme.Provider = &Provider{}

// NewProvider returns a provider for the Package Control.
// Search only in the informed labels.
func NewProvider(labels []string, opts ...Option) *Provider {
	p := &Provider{
		labels: labels,
		cli:    http.DefaultClient,
		tiker:  time.NewTicker(defaultRequestsInterval),
		logger: logrus.New(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type Option func(p *Provider)

func WithHTTPClient(cli *http.Client) Option {
	return func(p *Provider) {
		p.cli = cli
	}
}

func WithLogger(logger *logrus.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}

// SetRequestsInterval sets the interval between requests.
func (p *Provider) SetRequestsInterval(interval time.Duration) {
	p.tiker.Reset(interval)
}

// GetGallery retorna a galeria de temas.
// Em caso de erro retorna erro e a galeria de melhor esforço.
func (p *Provider) GetGallery() (theme.Gallery, error) {
	log := p.operation("Provider.GetGallery")
	names, err := p.fetchPackagesNames()
	if err != nil {
		log.WithError(err).Error("error on fetch package names")
		return nil, err
	}

	pkgs, err := p.fetchPackages(names)
	if err != nil {
		log.WithError(err).Error("error on fetch packages")
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
			Name:          pkg.Name,
			Description:   pkg.Description,
			Author:        strings.Join(pkg.Authors, ", "),
			Provider:      providerName,
			Readme:        pkg.Readme,
			ProjectRepoID: common.Hash(srcrepo),
			ProjectRepo:   srcrepo,
			LastUpdate:    pkg.LastModified,
		}

		if len(pkg.Versions) > 0 {
			th.Version = strconv.Itoa(pkg.Versions[len(pkg.Versions)-1])
		}

		gallery = append(gallery, th)
	}
	return gallery, nil
}

func (p *Provider) operation(operation string) *logrus.Entry {
	return p.logger.WithField("operation", operation)
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

func (p *Provider) fetchPackagesNames() ([]string, error) {
	log := p.operation("Provider.fetchPakcagesNames")
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	var (
		group errgroup.Group
		mux   sync.Mutex
	)

	var pkgnames []string

	for _, label := range p.labels {
		label := label
		select {
		case <-p.tiker.C:
			group.Go(func() error {
				log.WithField("label", label).Print("fetching label")

				resp, err := p.cli.Get(labelEndpoint + url.PathEscape(label) + ".json")
				if err != nil {
					log.WithError(err).WithField("label", label).Error("error on request label")
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.WithField("statusCode", resp.StatusCode).Error("respose status error")
					return fmt.Errorf("error on fetch packages for label %s", label)
				}

				parsed, err := parsePackageNames(resp.Body)
				if err != nil {
					log.WithError(err).Error("error on parse package names")
					return err
				}

				mux.Lock()
				pkgnames = append(pkgnames, parsed...)
				mux.Unlock()

				return nil
			})
		}
	}

	err := group.Wait()
	if err != nil {
		return pkgnames, err
	}

	return pkgnames, nil
}

func (p *Provider) fetchPackages(pkgnames []string) ([]pkg, error) {
	log := p.operation("Provider.fetchPackages")
	if p.cli == nil {
		return nil, errors.New("the http client must be specified")
	}

	var (
		mux   sync.Mutex
		group errgroup.Group
	)

	var pkgs []pkg

	for _, pkgname := range pkgnames {
		pkname := pkgname

		select {
		case <-p.tiker.C:
			group.Go(func() error {
				resp, err := p.cli.Get(packageEndpoint + url.PathEscape(pkname) + ".json")
				if err != nil {
					log.WithError(err).WithField("package", pkname).Error("error on request package")
					return nil
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					if resp.StatusCode == http.StatusNotFound { // Pacote não existe mais.
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
					log.Error("response data format not accepted")
					return fmt.Errorf("not json data returned for the %s package", pkname)
				}

				pk, err := parsePackage(resp.Body)
				if err != nil {
					log.WithError(err).Error("error on parse packages from response body")
					return err
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
