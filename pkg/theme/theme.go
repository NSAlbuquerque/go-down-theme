package theme

import (
	"io"
	"time"
)

// Theme metainformações sobre temas
type Theme struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Light       bool      `json:"light"`
	ProjectRepo string    `json:"projectRepo"`
	Version     string    `json:"version,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// Gallery representa um conjunto de metainformações sobre temas.
type Gallery []Theme

// Provider disponibiliza metainformações sobre temas a partir de uma fonte.
type Provider interface {
	GetGallery() (Gallery, error)
	Download(url string, to io.Writer) error
}
