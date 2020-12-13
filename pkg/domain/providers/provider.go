package providers

import (
	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
)

// Provider seeks theme metadata from a source.
type Provider interface {
	GetGallery() (themes.Gallery, error)
}
