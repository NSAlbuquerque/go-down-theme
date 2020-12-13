package colorsubl

import (
	"testing"

	"github.com/albuquerq/go-down-theme/pkg/domain/providers"

	"github.com/stretchr/testify/assert"
)

func Test_ColorSublProvider_GetGallery(t *testing.T) {
	tests := []struct {
		name     string
		provider providers.Provider
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

			assert.Greater(t, len(gallery), 0)
			for _, theme := range gallery {
				assert.NotEmpty(t, theme.URL)
				t.Log(theme.Name, theme.URL)
			}
		})
	}
}
