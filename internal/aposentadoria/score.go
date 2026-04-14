package aposentadoria

import "time"

const (
	// ScoreJudicial é o score atribuído a processos judicializados.
	ScoreJudicial = 9

	// ScorePrioridade é o score atribuído a processos com prioridade aprovada.
	ScorePrioridade = 8
)

func CalculateAge(t time.Time) int {
	now := time.Now()
	age := now.Year() - t.Year()

	if now.Month() < t.Month() ||
		(now.Month() == t.Month() && now.Day() < t.Day()) {
		age--
	}

	return age
}

// CalculateScore calcula o score de um processo de aposentadoria.
//
// Se o processo for judicializado, retorna [ScoreJudicial].
// Se o processo possuir prioridade aprovada, retorna [ScorePrioridade].
//
// Caso contrário, calcula normalmente:
//   - Se a idade é maior ou igual a 60, adiciona um ponto.
//   - Se a idade é maior ou igual a 80, adiciona dois pontos.
//   - Se o requerente possui doença grave ou invalidez, adiciona dois pontos.
func CalculateScore(dataNasc time.Time, invalidez, judicial, prioridade bool) int {
	if judicial {
		return ScoreJudicial
	}

	if prioridade {
		return ScorePrioridade
	}

	score := 0

	age := CalculateAge(dataNasc)
	if age >= 60 {
		score += 1
	}
	if age >= 80 {
		score += 2
	}
	if invalidez {
		score += 2
	}

	return score
}
