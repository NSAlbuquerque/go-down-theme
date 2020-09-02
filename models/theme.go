package models

// ThemeMetadata metainformações sobre temas
type ThemeMetadata struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Light       bool   `json:"light"`
	ProjectRepo string `json:"projectRepo"`
}

// Gallery representa um conjunto de metainformações sobre temas.
type Gallery []ThemeMetadata
