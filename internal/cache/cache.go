package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCacheMiss é o erro retornado quando alguma chave (key) não é encontrada no cache.
	ErrCacheMiss = errors.New("cache miss")
)

// Cache define uma interface mínima para operações de armazenamento temporário de dados.
// Inspirado pela implementação de cache do Laravel: https://laravel.com/docs/12.x/cache.
type Cache interface {
	// Get retorna os dados do cache pertencentes à chave informada.
	// Retorna [ErrCacheMiss] quando a chave não estiver presente no cache.
	Get(ctx context.Context, key string) ([]byte, error)
	// Put adiciona a chave e valor no cache pelo tempo definido por ttl.
	// Se ttl for zero, a chave não possui um tempo de expiração.
	Put(ctx context.Context, key string, data []byte, ttl time.Duration) error
	// Del remove o valor vinculado à uma chave do cache. Se a chave não existir,
	// a operação é um noop.
	Del(ctx context.Context, key string) error
	// Remember combina os métodos Put e Get. Se a chave não existir, executa fn para obter
	// os dados e salva no cache com o ttl informado.
	Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error)
}
