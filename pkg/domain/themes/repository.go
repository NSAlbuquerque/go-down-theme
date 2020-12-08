package themes

import "context"

// Repository of themes.
type Repository interface {
	Save(context.Context, *Theme) error
	SaveGallery(context.Context, Gallery) error
	Get(context.Context, string) (*Theme, error)
	List(context.Context) (Gallery, error)
	Delete(context.Context, string) error
}
