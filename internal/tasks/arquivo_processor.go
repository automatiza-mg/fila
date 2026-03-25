package tasks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/markdown"
)

type ArquivoProcessor struct {
	store   *database.Store
	storage blob.Storage
	cv      *docintel.AzureDocIntel
}

func NewArquivoProcessor(store *database.Store, storage blob.Storage, cv *docintel.AzureDocIntel) *ArquivoProcessor {
	return &ArquivoProcessor{
		store:   store,
		storage: storage,
		cv:      cv,
	}
}

// extractContent extrai o texto de um documento de acordo com o content-type.
// Para HTML, converte para markdown. Para outros formatos, utiliza a Azure Document Intelligence.
func (ap *ArquivoProcessor) extractContent(ctx context.Context, body []byte, contentType string) (string, string, error) {
	if markdown.IsHTML(contentType) {
		md, err := markdown.ConvertHTML(bytes.NewReader(body), contentType, markdown.WithoutImg())
		if err != nil {
			return "", "", err
		}
		return md, "markdown", nil
	}

	text, err := ap.cv.ExtractText(ctx, bytes.NewReader(body), contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract text: %w", err)
	}
	return text, "plain", nil
}

// Process cria um novo [database.Arquivo] com base no [io.Reader] e Content-Type
// informados.
func (ap *ArquivoProcessor) Process(ctx context.Context, r io.Reader, contentType string) (*database.Arquivo, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler conteúdo do arquivo: %w", err)
	}

	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])

	arq, err := ap.store.GetArquivo(ctx, hash)
	if err == nil {
		return arq, nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	storageKey := fmt.Sprintf("arquivos/%s", hash)

	err = ap.storage.Put(ctx, storageKey, bytes.NewReader(body), contentType)
	if err != nil {
		return nil, fmt.Errorf("falha ao armazenar arquivo: %w", err)
	}

	conteudo, formato, err := ap.extractContent(ctx, body, contentType)
	if err != nil {
		return nil, err
	}

	arq = &database.Arquivo{
		Hash:            hash,
		ChaveStorage:    storageKey,
		ContentType:     contentType,
		Conteudo:        conteudo,
		FormatoConteudo: formato,
	}

	err = ap.store.SaveArquivo(ctx, arq)
	if err != nil {
		return nil, fmt.Errorf("falha ao salvar arquivo: %w", err)
	}

	return arq, nil
}
