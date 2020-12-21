package providers

import (
	"log"
	"sync"

	"github.com/albuquerq/go-down-theme/pkg/domain/vos"
)

var (
	providers map[vos.ProviderName]Provider
	mux       sync.Mutex
)

// Register theme provider implementation.
func Register(name vos.ProviderName, provider Provider) {
	mux.Lock()
	defer mux.Unlock()
	providers[name] = provider
}

// Get returns a registred provider by name.
func Get(name vos.ProviderName) Provider {
	if p, exists := providers[name]; exists {
		return p
	}
	log.Fatalf("%s provider not registred", name)
	return nil
}

// List returns registered providers.
func List() []Provider {
	var provids []Provider

	for _, p := range providers {
		provids = append(provids, p)
	}
	return provids
}
