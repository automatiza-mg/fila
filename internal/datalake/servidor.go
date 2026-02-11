package datalake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Servidor struct {
	IDPessoa          int64     `json:"id_pessoa"`
	Nome              string    `json:"nome"`
	Masp              string    `json:"masp"`
	CPF               string    `json:"cpf"`
	Sexo              string    `json:"sexo"`
	DataNascimento    time.Time `json:"data_nascimento"`
	PossuiDeficiencia bool      `json:"possui_deficiencia"`
}

// GetServidorByCPF retorna os dados básicos do servidor pelo cpf. Retorna [ErrNotFound] caso nenhum dado tenha
// sido encontrado.
//
// O valor do cpf deve ser informado sem formatação, contendo apenas números como: '12345678910'.
func (d *DataLake) GetServidorByCPF(ctx context.Context, cpf string) (*Servidor, error) {
	q := `
	SELECT 
		CAST(pfj.ddnrpessoafisjur AS INT),
		TRIM(pfj.ddnomepessoa),
		CAST(cp.ddnrmaspfunc AS STRING) || '-' || cp.dddvmaspfunc,
		TRIM(doc.ddnrdocumento),
		pf.ddcdsexo,
		pf.dddtnascimento,
		cp.ddcddeficiencia <> 0
	FROM db_dlseplag_prod_sisap_staging.tb_tbcomplpessoa cp
	JOIN db_dlseplag_prod_sisap_staging.tb_tbpessoa_fis_jur pfj ON cp.ddnrpessoafisjur = pfj.ddnrpessoafisjur
	JOIN db_dlseplag_prod_sisap_staging.tb_tbpessoa pf ON cp.ddnrpessoafisjur = pf.ddnrpessoafisjur 
	JOIN db_dlseplag_prod_sisap_staging.tb_tbdocumento doc ON cp.ddnrpessoafisjur = doc.ddnrpessoafisjur 
	JOIN db_dlseplag_prod_sisap_staging.tb_txtipodocumento tdoc ON doc.ddcdtipodocumento = tdoc.ddcdtipodocumento 
	WHERE TRIM(doc.ddnrdocumento) = ? AND doc.ddcdtipodocumento = 8`

	var servidor Servidor
	err := d.db.QueryRow(q, cpf).Scan(
		&servidor.IDPessoa,
		&servidor.Nome,
		&servidor.Masp,
		&servidor.CPF,
		&servidor.Sexo,
		&servidor.DataNascimento,
		&servidor.PossuiDeficiencia,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to lookup servidor: %w", err)
		}
	}

	return &servidor, nil
}
