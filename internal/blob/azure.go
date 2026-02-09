package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

var _ Storage = (*AzureStorage)(nil)

// Implementação de [Storage] de objetos usando o Azure Blob Storage.
//
// Referência: https://learn.microsoft.com/pt-br/azure/storage/blobs/storage-blobs-introduction
type AzureStorage struct {
	cfg        *Config
	serviceURL string
	client     *azblob.Client
}

// Constrói a URL de serviço da Azure Blob Storage.
func AzureServiceURL(accountName string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
}

func NewAzureStorage(ctx context.Context, cfg *Config) (*AzureStorage, error) {
	serviceURL := AzureServiceURL(cfg.AzureAccount)

	var client *azblob.Client
	if cfg.AzureApiKey != "" {
		// Api Key
		cred, err := azblob.NewSharedKeyCredential(cfg.AzureAccount, cfg.AzureApiKey)
		if err != nil {
			return nil, err
		}
		client, err = azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, err
		}

	} else {
		// Fallback to environment
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, err
		}
	}

	return &AzureStorage{
		cfg:        cfg,
		serviceURL: serviceURL,
		client:     client,
	}, nil
}

func (s *AzureStorage) Put(ctx context.Context, key string, r io.Reader, contentType string) error {
	_, err := s.client.UploadStream(ctx, s.cfg.AzureContainer, key, r, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
		Concurrency: 3,
	})
	if err != nil {
		return fmt.Errorf("upload failed for key: %s: %w", key, err)
	}

	return nil
}

func (s *AzureStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	rc, err := s.client.DownloadStream(ctx, s.cfg.AzureContainer, key, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("download failed for key %s: %w", key, err)
	}

	return rc.Body, nil
}

func (s *AzureStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteBlob(ctx, s.cfg.AzureContainer, key, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil
		}
		return fmt.Errorf("delete failed for key %s: %w", key, err)
	}
	return nil
}

// Close é um NOP para a implementação de [Storage] da Azure.
func (s *AzureStorage) Close() error {
	return nil
}
