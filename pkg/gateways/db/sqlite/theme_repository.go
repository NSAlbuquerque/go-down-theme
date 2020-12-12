package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/albuquerq/go-down-theme/pkg/domain/themes"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = fmt.Errorf("%w", errors.New("err not found"))
)

// ThemeRepository for the SQLite3 database.
type ThemeRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

var _ themes.Repository = &ThemeRepository{}

// NewThemeRepository returns a theme repository for the SQLite3 database.
func NewThemeRepository(db *sql.DB, opts ...Option) *ThemeRepository {
	r := &ThemeRepository{
		logger: logrus.New(),
		db:     db,
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Option apply depedencies to the theme repository.
type Option func(r *ThemeRepository)

// WithLogger applies custom logger to the repository.
func WithLogger(logger *logrus.Logger) Option {
	return func(r *ThemeRepository) {
		r.logger = logger
	}
}

// Save stores a theme in the database.
func (r *ThemeRepository) Save(ctx context.Context, t *themes.Theme) error {
	log := r.operation("ThemeRepository.Save")
	const cmd = `
		INSERT INTO themes(
			id,
			name,
			author,
			description,
			url,
			light,
			project_repo,
			readme,
			version,
			license,
			provider,
			updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	id := uuid.New().String()
	_, err := r.db.ExecContext(
		ctx,
		cmd,
		id,
		t.Name,
		t.Author,
		t.Description,
		t.URL,
		t.Light,
		t.ProjectRepo,
		t.Readme,
		t.Version,
		t.License,
		t.Provider,
		t.LastUpdate,
	)
	if err != nil {
		log.WithError(err).Error("error on execute insert command")
		return err
	}

	t.ID = id

	return nil
}

// SaveThemes stores multiple themes at once.
func (r *ThemeRepository) SaveThemes(ctx context.Context, ths ...*themes.Theme) error {
	return r.save(ctx, "Repository.SaveThemes", ths...)
}

func (r *ThemeRepository) save(ctx context.Context, op string, ths ...*themes.Theme) error {
	log := r.operation(op)

	var (
		sb   strings.Builder
		args []interface{}
	)

	sb.WriteString(`
		INSERT INTO themes(
			id,
			name,
			author,
			description,
			url,
			light,
			project_repo,
			readme,
			version,
			license,
			provider,
			last_update)
		VALUES `)

	qtd := len(ths)
	for i := 0; i < qtd; i++ {
		sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if i < qtd-1 {
			sb.WriteString(",\n")
		}
		ths[i].ID = uuid.New().String()
		args = append(
			args,
			ths[i].ID,
			ths[i].Name,
			ths[i].Author,
			ths[i].Description,
			ths[i].URL,
			ths[i].Light,
			ths[i].ProjectRepo,
			ths[i].Readme,
			ths[i].Version,
			ths[i].License,
			string(ths[i].Provider),
			ths[i].LastUpdate,
		)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.WithError(err).Error("error on creating database transaction")
		return err
	}

	result, err := tx.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		log.WithError(err).Error("error on execute insert command")
		clearIDs(ths...)
		tx.Rollback()
		return err
	}
	if inserts, err := result.RowsAffected(); err != nil || inserts != int64(qtd) {
		log.Error("missing insert theme")
		clearIDs(ths...)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("error on commit database inserts")
		clearIDs(ths...)
		return err
	}

	return nil
}

func clearIDs(ths ...*themes.Theme) {
	for i := 0; i < len(ths); i++ {
		ths[i].ID = ""
	}
}

// Get returns a theme by id.
func (r *ThemeRepository) Get(ctx context.Context, id string) (*themes.Theme, error) {
	return r.getWhere(ctx, "ThemeRepository.Get", "WHERE id = ?", id)
}

// Delete removes a theme by id.
func (r *ThemeRepository) Delete(ctx context.Context, id string) error {
	log := r.operation("ThemeRepository.Delete").WithField("theme_id", id)

	const cmd = `DELETE FROM themes WHERE id = ?`

	result, err := r.db.ExecContext(ctx, cmd, id)
	if err != nil {
		log.WithError(err).Error("error on execute delete command")
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		log.WithError(err).Error("error on query affected rows number")
		return err
	}
	if n == 0 {
		log.Print("theme not found")
		return ErrNotFound
	}

	return nil
}

// List all themes stored.
func (r *ThemeRepository) List(ctx context.Context, filter *themes.ListFilter) (themes.Gallery, error) {
	var ops string

	if filter != nil {
		if filter.Limit > 0 {
			ops += fmt.Sprintf(" LIMIT %d", filter.Limit)
		}
		if filter.Offset > 0 {
			ops += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}
	return r.listWhere(ctx, "ThemeRepository.List", "ORDER BY name "+ops)
}

func (r *ThemeRepository) operation(op string) *logrus.Entry {
	return r.logger.WithField("operation", op)
}

func (r *ThemeRepository) getWhere(ctx context.Context, operation, where string, args ...interface{}) (*themes.Theme, error) {
	log := r.operation(operation)
	const cmd = `
		SELECT 
			id,
			name,
			author,
			description,
			url,
			light,
			project_repo,
			readme,
			version,
			license,
			provider,
			last_update
		FROM themes %s
	`
	var (
		t          themes.Theme
		lastUpdate sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(cmd, where), args...).Scan(
		&t.ID,
		&t.Name,
		&t.Author,
		&t.Description,
		&t.URL,
		&t.Light,
		&t.ProjectRepo,
		&t.Readme,
		&t.Version,
		&t.License,
		&t.Provider,
		&lastUpdate,
	)
	if err != nil {
		log.WithError(err).Error("error on execute query comand")
		return nil, err
	}

	if lastUpdate.Valid {
		t.LastUpdate = lastUpdate.Time
	}

	return &t, nil
}

func (r *ThemeRepository) listWhere(ctx context.Context, operation, where string, args ...interface{}) (themes.Gallery, error) {
	log := r.operation(operation)
	const cmd = `
		SELECT
			id, 
			name,
			author,
			description,
			url,
			light,
			project_repo,
			readme,
			version,
			license,
			provider,
			last_update
		FROM themes %s
	`
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(cmd, where), args...)
	if err != nil {
		log.WithError(err).Error("error on execute query command")
		return nil, err
	}
	defer rows.Close()

	var galery themes.Gallery

	for rows.Next() {
		var (
			t          themes.Theme
			lastUpdate sql.NullTime
		)
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Author,
			&t.Description,
			&t.URL,
			&t.Light,
			&t.ProjectRepo,
			&t.Readme,
			&t.Version,
			&t.License,
			&t.Provider,
			&lastUpdate,
		)
		if err != nil {
			log.WithError(err).Error("error on scan query values")
			return nil, err
		}

		if lastUpdate.Valid {
			t.LastUpdate = lastUpdate.Time
		}

		galery = append(galery, t)
	}

	return galery, nil
}
