package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Arquivo struct {
	Hash            string    `db:"hash"`
	ChaveStorage    string    `db:"chave_storage"`
	ContentType     string    `db:"content_type"`
	Conteudo        string    `db:"conteudo"`
	FormatoConteudo string    `db:"formato_conteudo"`
	CriadoEm        time.Time `db:"criado_em"`
}

// SaveArquivo insere um novo arquivo no banco de dados
func (s *Store) SaveArquivo(ctx context.Context, a *Arquivo) error {
	q := `
	INSERT INTO arquivos (hash, chave_storage, content_type, conteudo, formato_conteudo)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (hash) DO NOTHING
	RETURNING criado_em`
	args := []any{
		a.Hash,
		a.ChaveStorage,
		a.ContentType,
		a.Conteudo,
		a.FormatoConteudo,
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
	SELECT hash, chave_storage, content_type, conteudo, formato_conteudo, criado_em
	FROM arquivos
	WHERE hash = $1`

	rows, err := s.db.Query(ctx, q, hash)
	if err != nil {
		return nil, err
	}
	a, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Arquivo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

// GetArquivosMap retorna um mapa de hash -> Arquivo para os hashes informados.
func (s *Store) GetArquivosMap(ctx context.Context, hashes []string) (map[string]*Arquivo, error) {
	if len(hashes) == 0 {
		return make(map[string]*Arquivo), nil
	}

	q := `
	SELECT hash, chave_storage, content_type, conteudo, formato_conteudo, criado_em
	FROM arquivos
	WHERE hash = ANY($1)`

	rows, err := s.db.Query(ctx, q, hashes)
	if err != nil {
		return nil, err
	}
	list, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Arquivo])
	if err != nil {
		return nil, err
	}
	arquivoMap := make(map[string]*Arquivo, len(list))
	for _, a := range list {
		arquivoMap[a.Hash] = a
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
