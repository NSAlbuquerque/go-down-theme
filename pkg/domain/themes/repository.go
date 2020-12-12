package themes

import "context"

// Repository of themes.
type Repository interface {
	Save(context.Context, *Theme) error
	SaveThemes(context.Context, ...*Theme) error
	Get(context.Context, string) (*Theme, error)
	List(context.Context, *ListFilter) (Gallery, error)
	Delete(context.Context, string) error
}

// ListFilter filter querys.
type ListFilter struct {
	Limit  int
	Offset int
}
