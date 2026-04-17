package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Processo struct {
	NumeroProcesso   string
	DataNascimento   time.Time
	DataRequerimento time.Time
	Invalidez        bool
	Judicial         bool
	Score            *int
}

func main() {
	postgresURL := flag.String("pg", "", "Define a conexão com o banco para cálculo do score")
	flag.Parse()

	pool, err := pgxpool.New(context.Background(), *postgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	q := `
	SELECT 
		numero_processo, data_nascimento, data_requerimento, invalidez, judicial, score
	FROM processos_aposentadoria_teste`

	rows, err := pool.Query(context.Background(), q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var processo Processo

		err := rows.Scan(
			&processo.NumeroProcesso,
			&processo.DataNascimento,
			&processo.DataRequerimento,
			&processo.Invalidez,
			&processo.Judicial,
			&processo.Score,
		)
		if err != nil {
			log.Fatal(err)
		}

		score := aposentadoria.CalculateScore(processo.DataNascimento, processo.Invalidez, processo.Judicial, false)

		q := `UPDATE processos_aposentadoria_teste SET score = $2 WHERE numero_processo = $1`
		_, err = pool.Exec(context.Background(), q, processo.NumeroProcesso, score)
		if err != nil {
			log.Printf("Failed to update: %v", err)
		}

		total++
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Total de %d processos atualizados", total)
}
