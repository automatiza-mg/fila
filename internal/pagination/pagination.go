package pagination

import (
	"net/http"
	"strconv"
)

// Params contém os parâmetros de paginação com valores padrão sensatos.
type Params struct {
	Page  int
	Limit int
}

// Constantes padrão de paginação.
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 50
	MinLimit     = 1
	MinPage      = 1
)

type Result[T any] struct {
	Data        []T  `json:"data"`
	Limit       int  `json:"limit"`
	CurrentPage int  `json:"current_page"`
	TotalCount  int  `json:"total_count"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// Offset calcula a quantidade de linhas que o banco de dados deve pular
// dado uma página e limite.
func Offset(page, limit int) int {
	return (page - 1) * limit
}

// NewResult cria um novo resultado paginado com os dados e informações fornecidos.
func NewResult[T any](data []T, page, totalCount, limit int) *Result[T] {
	totalPages := (totalCount + limit - 1) / limit

	return &Result[T]{
		Data:        data,
		Limit:       limit,
		CurrentPage: page,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}
}

// ParseQuery extrai e valida os parâmetros de paginação da query string HTTP.
// É tolerante - valores inválidos usam valores padrão sensatos em vez de retornar erros.
//
// Parâmetros da query:
//   - page: número da página (baseado em 1), padrão é 1
//   - limit: número de itens por página, padrão é 20, limitado a 100
//
// Exemplos:
//   - ?page=2&limit=50 → Params{Page: 2, Limit: 50}
//   - ?page=abc&limit=50 → Params{Page: 1, Limit: 50} (página inválida usa padrão)
//   - ?page=2 → Params{Page: 2, Limit: 20} (limite ausente usa padrão)
//   - ?page=-5&limit=200 → Params{Page: 1, Limit: 100} (limitado ao intervalo válido)
func ParseQuery(r *http.Request) Params {
	params := Params{
		Page:  DefaultPage,
		Limit: DefaultLimit,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = max(MinPage, page)
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = min(max(MinLimit, limit), MaxLimit)
		}
	}

	return params
}
