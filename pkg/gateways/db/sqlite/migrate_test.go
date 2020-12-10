package sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	err = Migrate("file:migrations", db, logrus.New())
	assert.NoError(t, err)
}
