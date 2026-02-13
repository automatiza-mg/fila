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
		name      string
		dataNasc  time.Time
		invalidez bool
		expected  int
	}{
		{name: "jovem sem invalidez", dataNasc: birthDate(30), invalidez: false, expected: 0},
		{name: "jovem com invalidez", dataNasc: birthDate(30), invalidez: true, expected: 2},
		{name: "60 anos sem invalidez", dataNasc: birthDate(60), invalidez: false, expected: 1},
		{name: "60 anos com invalidez", dataNasc: birthDate(70), invalidez: true, expected: 3},
		{name: "80 anos sem invalidez", dataNasc: birthDate(80), invalidez: false, expected: 3},
		{name: "80 anos com invalidez", dataNasc: birthDate(85), invalidez: true, expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScore(tt.dataNasc, tt.invalidez)
			if got != tt.expected {
				t.Fatalf("CalculateScore(%v, %v) = %d, expected %d", tt.dataNasc, tt.invalidez, got, tt.expected)
			}
		})
	}
}
