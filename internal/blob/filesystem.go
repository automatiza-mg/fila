package blob

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Implementação mínima de [Storage] para uso em desenvolvimento.
type FilesystemStore struct {
	root *os.Root
}

func NewFilesystemStore(rootDir string) (*FilesystemStore, error) {
	if rootDir != "" {
		err := os.MkdirAll(rootDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}

	return &FilesystemStore{root: root}, nil
}

// Put adiciona um objeto ao sistema de arquivos local. O parâmetro contentType é ignorado nessa implementação.
func (s *FilesystemStore) Put(ctx context.Context, key string, r io.Reader, _ string) error {
	if dir := filepath.Dir(key); dir != "." {
		err := s.root.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	f, err := s.root.Create(key)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return nil
}

// Get retorna uma stream do objeto no sistema de arquivos local.
func (s *FilesystemStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	f, err := s.root.Open(key)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return f, nil
}

// Delete remove o objeto da sistema de arquivos local.
func (s *FilesystemStore) Delete(ctx context.Context, key string) error {
	err := s.root.RemoveAll(key)
	if err != nil {
		return err
	}
	return nil
}

// Close fecha o [os.Root] subjacente dessa implementação.
func (s *FilesystemStore) Close() error {
	return s.root.Close()
}
