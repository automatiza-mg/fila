package aposentadoria

import "time"

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
// Se a idade é maior ou igual a 60, adiciona um ponto.
// Se a idade é maior ou igual a 80, adiciona dois pontos.
// Se o requerente possui doença grave ou invalidez, adiciona dois pontos.
func CalculateScore(dataNasc time.Time, invalidez bool) int {
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
