package github

import "fmt"

// Repo represents a theme repository.
type Repo struct {
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Branch string `json:"branch"`
}

func (r Repo) String() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

// File ...
type File struct {
	Name        string
	DownloadURL string
}
