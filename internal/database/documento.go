package database

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Documento struct {
	ID           int64
	Numero       string
	ProcessoID   uuid.UUID
	Tipo         string
	Unidade      string
	LinkAcesso   string
	ContentType  string
	ChaveStorage string
	OCR          string
	MetadadosAPI json.RawMessage
	CriadoEm     time.Time
	AtualizadoEm time.Time
}

func (s *Store) SaveDocumento(ctx context.Context, d *Documento) error {
	q := `
	INSERT INTO documentos (
		numero, processo_id, tipo, unidade,
		link_acesso, content_type, chave_storage, ocr,
		metadados_api
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, criado_em, atualizado_em`
	args := []any{
		d.Numero,
		d.ProcessoID,
		d.Tipo,
		d.Unidade,
		d.LinkAcesso,
		d.ContentType,
		d.ChaveStorage,
		d.OCR,
		d.MetadadosAPI,
	}

	err := s.db.QueryRow(ctx, q, args...).Scan(&d.ID, &d.CriadoEm, &d.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetDocumento(ctx context.Context, id int64) (*Documento, error) {
	q := `
	SELECT
		id, numero, processo_id, tipo, unidade,
		link_acesso, content_type, chave_storage, ocr, metadados_api,
		criado_em, atualizado_em
	FROM documentos
	WHERE id = $1`

	var d Documento
	err := s.db.QueryRow(ctx, q, id).Scan(
		&d.ID,
		&d.Numero,
		&d.ProcessoID,
		&d.Tipo,
		&d.Unidade,
		&d.LinkAcesso,
		&d.ContentType,
		&d.ChaveStorage,
		&d.OCR,
		&d.MetadadosAPI,
		&d.CriadoEm,
		&d.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (s *Store) GetDocumentoByNumero(ctx context.Context, numero string) (*Documento, error) {
	q := `
	SELECT
		id, numero, processo_id, tipo, unidade,
		link_acesso, content_type, chave_storage, ocr, metadados_api,
		criado_em, atualizado_em
	FROM documentos
	WHERE numero = $1`

	var d Documento
	err := s.db.QueryRow(ctx, q, numero).Scan(
		&d.ID,
		&d.Numero,
		&d.ProcessoID,
		&d.Tipo,
		&d.Unidade,
		&d.LinkAcesso,
		&d.ContentType,
		&d.ChaveStorage,
		&d.OCR,
		&d.MetadadosAPI,
		&d.CriadoEm,
		&d.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (s *Store) ListDocumentos(ctx context.Context, processoID uuid.UUID) ([]*Documento, error) {
	q := `
	SELECT
		id, numero, processo_id, tipo, unidade,
		link_acesso, content_type, chave_storage, ocr, metadados_api,
		criado_em, atualizado_em
	FROM documentos
	WHERE processo_id = $1`

	rows, err := s.db.Query(ctx, q, processoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dd := make([]*Documento, 0)
	for rows.Next() {
		var d Documento
		err := rows.Scan(
			&d.ID,
			&d.Numero,
			&d.ProcessoID,
			&d.Tipo,
			&d.Unidade,
			&d.LinkAcesso,
			&d.ContentType,
			&d.ChaveStorage,
			&d.OCR,
			&d.MetadadosAPI,
			&d.CriadoEm,
			&d.AtualizadoEm,
		)
		if err != nil {
			return nil, err
		}
		dd = append(dd, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dd, nil
}

func (s *Store) GetDocumentosMap(ctx context.Context, processoIDs []uuid.UUID) (map[uuid.UUID][]*Documento, error) {
	if len(processoIDs) == 0 {
		return make(map[uuid.UUID][]*Documento), nil
	}

	q := `
	SELECT
		id, numero, processo_id, tipo, unidade,
		link_acesso, content_type, chave_storage, ocr, metadados_api,
		criado_em, atualizado_em
	FROM documentos
	WHERE processo_id = ANY($1)`

	rows, err := s.db.Query(ctx, q, processoIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docMap := make(map[uuid.UUID][]*Documento, len(processoIDs))
	for rows.Next() {
		var d Documento
		err := rows.Scan(
			&d.ID,
			&d.Numero,
			&d.ProcessoID,
			&d.Tipo,
			&d.Unidade,
			&d.LinkAcesso,
			&d.ContentType,
			&d.ChaveStorage,
			&d.OCR,
			&d.MetadadosAPI,
			&d.CriadoEm,
			&d.AtualizadoEm,
		)
		if err != nil {
			return nil, err
		}

		docMap[d.ProcessoID] = append(docMap[d.ProcessoID], &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docMap, nil
}
