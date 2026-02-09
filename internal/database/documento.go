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
		metadados_api, criado_em, atualizado_em
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id, criado_em, atualizado_em`
	args := []any{
		d.Numero, d.ProcessoID, d.Tipo, d.Unidade,
		d.LinkAcesso, d.ContentType, d.ChaveStorage, d.OCR,
		d.MetadadosAPI, d.CriadoEm, d.AtualizadoEm,
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
