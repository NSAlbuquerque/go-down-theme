package theme

import (
	"time"
)

// Theme metainformações sobre temas
type Theme struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Light       bool   `json:"light"`
	ProjectRepo string `json:"projectRepo"`
	Readme      string `json:"readme"`

	Version   string    `json:"version,omitempty"`
	Licence   string    `json:"licence,omitempty"`
	Provider  string    `json:"provider,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// Gallery representa um conjunto de metainformações sobre temas.
type Gallery []Theme

// Provider disponibiliza metainformações sobre temas a partir de uma fonte.
type Provider interface {
	GetGallery() (Gallery, error)
}
