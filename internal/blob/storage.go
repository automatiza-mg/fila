package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrNotFound é o erro retornado quando a chave informada em [Storage.Get] não existe.
	ErrNotFound = errors.New("object not found")
)

// Storage é uma interface mínima para o armazendo de objetos (blob storage).
type Storage interface {
	// Get retorna um [io.ReadCloser] para a chave informada. Caso a chave não seja encontrada, as implementações
	// devem retornar [ErrNotFound].
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Put adiciona um objeto à storage com a chave e metadados informados. Em caso de conflito, o valor deve
	// ser sobscrito.
	Put(ctx context.Context, key string, r io.Reader, contentType string) error
	// Delete remove um objeto da storage. Implementações não devem retornar erro no caso de chaves não encontradas.
	Delete(ctx context.Context, key string) error
	// Close fecha os recursos subjacentes da storage, se necessário.
	Close() error
}

// New retorna uma [Storage] de acordo com a configuração fornecida.
func New(ctx context.Context, cfg *Config) (Storage, error) {
	switch cfg.Provider {
	case "filesystem":
		return NewFilesystemStore(cfg.FilesystemRoot)
	case "azure":
		return NewAzureStorage(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown storage provider: %q", cfg.Provider)
	}
}
