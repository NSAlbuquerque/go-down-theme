package themes

import "context"

// Repository of themes.
type Repository interface {
	Save(context.Context, *Theme) error
	SaveGallery(context.Context, Gallery) error
	Get(context.Context, string) (*Theme, error)
	List(context.Context, *ListFilter) (Gallery, error)
	Delete(context.Context, string) error
}

// ListFilter filter querys.
type ListFilter struct {
	Limit  int
	Offset int
}
