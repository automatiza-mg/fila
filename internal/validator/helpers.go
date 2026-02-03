package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// Fonte: https://html.spec.whatwg.org/#valid-e-mail-address
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	CpfRX   = regexp.MustCompile(`^\d{11}$|^\d{3}\.\d{3}\.\d{3}\-\d{2}$`)
)

// Verifica se a string s é compativel com a expressão regular rx.
func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// Verifica se a string s não é vazia.
func NotBlank(s string) bool {
	s = strings.TrimSpace(s)
	return utf8.RuneCountInString(s) > 0
}

// Verifica se a string s possui tamanho maior ou igual a n.
func MinLength(s string, n int) bool {
	return utf8.RuneCountInString(s) >= n
}

// Verifica se a string s possui tamanho menor ou igual a n.
func MaxLength(s string, n int) bool {
	return utf8.RuneCountInString(s) <= n
}

// Verifica se a string s possui tamanho igual a n.
func Length(s string, n int) bool {
	return utf8.RuneCountInString(s) == n
}

// Verifica se a slices values contém apenas valores únicos.
func Unique[T comparable](values []T) bool {
	unique := make(map[T]struct{})
	for _, value := range values {
		unique[value] = struct{}{}
	}
	return len(unique) == len(values)
}

// Reporta se o valor está na lista de valores permitidos.
func PermittedValue[T comparable](value T, permitted ...T) bool {
	return slices.Contains(permitted, value)
}

// Verifica se a string s é uma senha forte.
func StrongPassword(s string) bool {
	if utf8.RuneCountInString(s) < 8 {
		return false
	}

	var (
		digit   bool
		special bool
	)

	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r), unicode.IsSymbol(r):
			special = true
		}

		if digit && special {
			return true
		}
	}

	return digit && special
}
