package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var fs embed.FS

// Up aplica todas as migrações ao banco de dados.
func Up(db *sql.DB) error {
	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("pgx"); err != nil {
		return err
	}
	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}
