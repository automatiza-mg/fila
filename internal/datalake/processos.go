package datalake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UnidadeGeradora struct {
	SiglaUnidade string `json:"sigla_unidade"`
	IDUnidade    string `json:"id_unidade"`
}

type Processo struct {
	NumeroProcesso  string          `json:"numero_processo"`
	SiglaUnidade    string          `json:"sigla_unidade"`
	DataRecebimento time.Time       `json:"data_recebimento"`
	UnidadeGeradora UnidadeGeradora `json:"unidade_geradora"`
}

func (d *DataLake) ListProcessosAbertos(ctx context.Context, unidade string) ([]Processo, int, error) {
	q := `
	SELECT
		numero_processo,
		sigla_unidade_andamento_processo,
		data_andamento_processo,
		CAST(id_unidade_geradora_processo AS INT),
		sigla_unidade_geradora_processo,
		COUNT(*) OVER()
	FROM (
		SELECT
			numero_processo,
			sigla_unidade_andamento_processo,
			data_andamento_processo,
			id_unidade_geradora_processo,
			sigla_unidade_geradora_processo,
			ROW_NUMBER() OVER (
				PARTITION BY numero_processo, sigla_unidade_andamento_processo
				ORDER BY data_andamento_processo ASC
			) AS rn
		FROM db_dlseplag_prod_dlsei_reporting.vw_sei_017_andamento_processo_aberto_automatiza
		WHERE sigla_unidade_andamento_processo = ?
	) AS t1
	WHERE t1.rn = 1`

	rows, err := d.db.Query(q, unidade)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	totalCount := 0
	processos := make([]Processo, 0)
	for rows.Next() {
		var p Processo
		err := rows.Scan(
			&p.NumeroProcesso,
			&p.SiglaUnidade,
			&p.DataRecebimento,
			&p.UnidadeGeradora.IDUnidade,
			&p.UnidadeGeradora.SiglaUnidade,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		processos = append(processos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return processos, totalCount, nil
}

func (d *DataLake) ListUnidadesDisponiveis(ctx context.Context) ([]string, error) {
	q := `
	SELECT DISTINCT sigla_unidade_andamento_processo
	FROM db_dlseplag_prod_dlsei_reporting.vw_sei_017_andamento_processo_aberto_automatiza`

	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unidades := make([]string, 0)
	for rows.Next() {
		var u string
		err := rows.Scan(&u)
		if err != nil {
			return nil, err
		}

		unidades = append(unidades, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return unidades, nil
}

// GetDataRecebimento retorna a data de recebimento de um processo pelo número e unidade.
func (d *DataLake) GetDataRecebimento(ctx context.Context, numero, unidade string) (time.Time, error) {
	q := `
	SELECT data_andamento_processo
	FROM (
		SELECT
			data_andamento_processo,
			ROW_NUMBER() OVER (
				PARTITION BY numero_processo, sigla_unidade_andamento_processo
				ORDER BY data_andamento_processo ASC
			) AS rn
		FROM db_dlseplag_prod_dlsei_reporting.vw_sei_017_andamento_processo_aberto_automatiza
		WHERE numero_processo = ?
		AND sigla_unidade_andamento_processo = ?
	) AS t1
	WHERE t1.rn = 1`

	var dataRecebimento time.Time
	err := d.db.QueryRowContext(ctx, q, numero, unidade).Scan(&dataRecebimento)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return time.Time{}, ErrNotFound
		default:
			return time.Time{}, fmt.Errorf("falha ao buscar data de recebimento: %w", err)
		}
	}

	return dataRecebimento, nil
}
