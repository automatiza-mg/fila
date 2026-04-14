package aposentadoria

import (
	"testing"
	"time"
)

func TestCalculateScore(t *testing.T) {
	now := time.Now()
	birthDate := func(age int) time.Time {
		return now.AddDate(-age, 0, 0)
	}

	tests := []struct {
		name       string
		dataNasc   time.Time
		invalidez  bool
		judicial   bool
		prioridade bool
		expected   int
	}{
		{name: "jovem sem invalidez", dataNasc: birthDate(30), expected: 0},
		{name: "jovem com invalidez", dataNasc: birthDate(30), invalidez: true, expected: 2},
		{name: "60 anos sem invalidez", dataNasc: birthDate(60), expected: 1},
		{name: "60 anos com invalidez", dataNasc: birthDate(70), invalidez: true, expected: 3},
		{name: "80 anos sem invalidez", dataNasc: birthDate(80), expected: 3},
		{name: "80 anos com invalidez", dataNasc: birthDate(85), invalidez: true, expected: 5},
		{name: "prioridade", dataNasc: birthDate(30), prioridade: true, expected: ScorePrioridade},
		{name: "prioridade com invalidez", dataNasc: birthDate(85), invalidez: true, prioridade: true, expected: ScorePrioridade},
		{name: "judicial", dataNasc: birthDate(85), invalidez: true, judicial: true, expected: ScoreJudicial},
		{name: "judicial tem precedencia sobre prioridade", dataNasc: birthDate(85), invalidez: true, judicial: true, prioridade: true, expected: ScoreJudicial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScore(tt.dataNasc, tt.invalidez, tt.judicial, tt.prioridade)
			if got != tt.expected {
				t.Fatalf("CalculateScore() = %d, expected %d", got, tt.expected)
			}
		})
	}
}
