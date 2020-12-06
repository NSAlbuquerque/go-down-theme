package themes

// Repository of themes.
type Repository interface {
	Save(*Theme) error
	Get(string) (*Theme, error)
	List() (Gallery, error)
	Delete(id string) error
}
