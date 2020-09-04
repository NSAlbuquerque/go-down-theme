package colorsubl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/albuquerq/go-down-theme/pkg/theme"
)

func Test_ColorSublProvider_GetGallery(t *testing.T) {
	tests := []struct {
		name     string
		provider theme.Provider
	}{
		// TODO: Add test cases.
		{
			name:     "Busca a galeria de temas",
			provider: NewProvider(),
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			gallery, err := tt.provider.GetGallery()
			assert.NoError(t, err)

			assert.NotEqual(t, 0, len(gallery))
			for _, theme := range gallery {
				assert.NotEmpty(t, theme.URL)
				t.Log(theme.Name, theme.URL)
			}
		})
	}
}
