package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Arquivo struct {
	Hash         string
	ChaveStorage string
	OCR          string
	ContentType  string
	CriadoEm     time.Time
}

// SaveArquivo insere um novo arquivo no banco de dados
func (s *Store) SaveArquivo(ctx context.Context, a *Arquivo) error {
	q := `
	INSERT INTO arquivos (hash, chave_storage, ocr, content_type)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (hash) DO NOTHING
	RETURNING criado_em`
	args := []any{
		a.Hash,
		a.ChaveStorage,
		a.OCR,
		a.ContentType,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(&a.CriadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

// GetArquivo retorna um arquivo pelo hash.
func (s *Store) GetArquivo(ctx context.Context, hash string) (*Arquivo, error) {
	q := `
	SELECT hash, chave_storage, ocr, content_type, criado_em
	FROM arquivos
	WHERE hash = $1`

	var a Arquivo
	err := s.db.QueryRow(ctx, q, hash).Scan(
		&a.Hash,
		&a.ChaveStorage,
		&a.OCR,
		&a.ContentType,
		&a.CriadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// GetArquivosMap retorna um mapa de hash -> Arquivo para os hashes informados.
func (s *Store) GetArquivosMap(ctx context.Context, hashes []string) (map[string]*Arquivo, error) {
	if len(hashes) == 0 {
		return make(map[string]*Arquivo), nil
	}

	q := `
	SELECT hash, chave_storage, ocr, content_type, criado_em
	FROM arquivos
	WHERE hash = ANY($1)`

	rows, err := s.db.Query(ctx, q, hashes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	arquivoMap := make(map[string]*Arquivo, len(hashes))
	for rows.Next() {
		var a Arquivo
		err := rows.Scan(
			&a.Hash,
			&a.ChaveStorage,
			&a.OCR,
			&a.ContentType,
			&a.CriadoEm,
		)
		if err != nil {
			return nil, err
		}
		arquivoMap[a.Hash] = &a
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return arquivoMap, nil
}

// DeleteArquivo remove um arquivo pelo hash.
func (s *Store) DeleteArquivo(ctx context.Context, hash string) error {
	q := `DELETE FROM arquivos WHERE hash = $1`
	_, err := s.db.Exec(ctx, q, hash)
	if err != nil {
		return err
	}
	return nil
}
