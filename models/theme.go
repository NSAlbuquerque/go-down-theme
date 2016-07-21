package models

// ThemeMetaData metainformações sobre temas
type ThemeMetaData struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Light bool   `json:"light"`
}

// Galery representa um conjunto de metainformações sobre temas.
type Galery []ThemeMetaData
