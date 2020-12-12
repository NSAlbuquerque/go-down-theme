package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDatabaseURL    = ":memory:"
	testMigrationsPath = "file://migrations"
)

func TestThemeRepository(t *testing.T) {

	log := logrus.New()

	db, err := sql.Open("sqlite3", testDatabaseURL)
	require.NoError(t, err)

	// removes tables
	_, err = db.ExecContext(context.Background(), "drop table if exists themes; drop table if exists schema_migrations;")
	require.NoError(t, err)

	err = Migrate(testMigrationsPath, db, log)
	require.NoError(t, err)

	repo := NewThemeRepository(db, WithLogger(log))

	t.Run("creates and get a theme", func(t *testing.T) {
		th := themes.Theme{
			Name:        "first theme",
			Description: "first test theme",
			Author:      "author",
			URL:         "http://github.com/albuquerq/teste",
		}

		err := repo.Save(context.Background(), &th)
		assert.NoError(t, err)

		th2, err := repo.Get(context.Background(), th.ID)
		assert.NoError(t, err)

		assert.Equal(t, th.ID, th2.ID)
		assert.Equal(t, th.Name, th2.Name)
		assert.Equal(t, th.Author, th2.Author)
	})

	t.Run("list all themes", func(t *testing.T) {
		ths, err := repo.List(context.Background(), nil)
		assert.NoError(t, err)
		assert.Greater(t, len(ths), 0)
		for i, th := range ths {
			assert.NotEmpty(t, th.Name)
			t.Logf("%d - %s", i+1, th.Name)
		}
	})

	t.Run("delete the fist theme found", func(t *testing.T) {
		ths, err := repo.List(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, ths[0])

		t.Logf("%s: %s", ths[0].ID, ths[0].Name)

		// Delete the fist time.
		err = repo.Delete(context.Background(), ths[0].ID)
		assert.NoError(t, err)

		// Do not delete the second time.
		err = repo.Delete(context.Background(), ths[0].ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("creates themes in batch", func(t *testing.T) {
		ths := []*themes.Theme{ // Ordened by name.
			{Name: "Theme 1"},
			{Name: "Theme 2"},
			{Name: "Theme 3"},
			{Name: "Theme 4"},
		}

		err := repo.SaveThemes(context.Background(), ths...)
		assert.NoError(t, err)

		ths2, err := repo.List(context.Background(), nil)
		assert.NoError(t, err)

		assert.Equal(t, len(ths), len(ths2))
		for i := 0; i < len(ths); i++ {
			t.Log(ths[i].Name)
			assert.Equal(t, ths[i].ID, ths2[i].ID)
			assert.Equal(t, ths[i].Name, ths2[i].Name)
		}
	})
}
