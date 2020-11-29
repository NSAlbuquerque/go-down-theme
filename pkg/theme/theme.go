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
	Hash        string `json:"hash"`
	Light       bool   `json:"light"`
	Version     string `json:"version,omitempty"`

	ProjectRepoID string    `json:"projectRepoId"`
	ProjectRepo   string    `json:"projectRepo"`
	Readme        string    `json:"readme"`
	License       string    `json:"license,omitempty"`
	Provider      string    `json:"provider,omitempty"`
	LastUpdate    time.Time `json:"updatedAt,omitempty"`
}

// Gallery representa um conjunto de metainformações sobre temas.
type Gallery []Theme

// Provider disponibiliza metainformações sobre temas a partir de uma fonte.
type Provider interface {
	GetGallery() (Gallery, error)
}
