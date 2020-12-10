package sqlite

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/sirupsen/logrus"
)

func Migrate(migrationsPath string, db *sql.DB, logger *logrus.Logger) error {
	log := logger.WithField("operation", "Migrate")

	dcfg, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.WithError(err).Error("error on config migrations")
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "godowntheme", dcfg)
	if err != nil {
		log.WithError(err).Error("error on creating migrations")
		return err
	}

	err = m.Up()
	if err != nil {
		log.WithError(err).Error("error on execute migrations")
		return err
	}
	return nil
}
