package tasks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/markdown"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/automatiza-mg/fila/internal/soap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"golang.org/x/sync/errgroup"
)

const (
	DownloadProcessoTimeout = 30 * time.Second
)

type DownloadProcessoArgs struct {
	ProcessoID uuid.UUID `json:"processo_id"`
}

func (args DownloadProcessoArgs) Kind() string {
	return "processo:download"
}

func (args DownloadProcessoArgs) KindAliases() []string {
	return []string{"pocesso:download"}
}

type DownloadProcessoWorker struct {
	pool    *pgxpool.Pool
	store   *database.Store
	storage blob.Storage
	sei     *sei.Client
	cv      *docintel.AzureDocIntel
	river.WorkerDefaults[DownloadProcessoArgs]
}

func NewDownloadProcessoWorker(pool *pgxpool.Pool, storage blob.Storage, sei *sei.Client, cv *docintel.AzureDocIntel) *DownloadProcessoWorker {
	return &DownloadProcessoWorker{
		pool:    pool,
		store:   database.New(pool),
		storage: storage,
		sei:     sei,
		cv:      cv,
	}
}

// Verifica se o Content-Type corresponde a um documento HTML.
func isHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html" || strings.HasSuffix(mediaType, "+html")
}

// Extrai e retorna o conteúdo e formato (plain / markdown) de um arquivo.
func (w *DownloadProcessoWorker) extractContent(ctx context.Context, body []byte, contentType string) (string, string, error) {
	if isHTML(contentType) {
		md, err := markdown.ConvertHTML(bytes.NewReader(body), contentType, markdown.WithoutImages())
		if err != nil {
			return "", "", nil
		}
		return md, "markdown", nil
	}

	text, err := w.cv.ExtractText(ctx, bytes.NewReader(body), contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract text: %w", err)
	}

	return text, "plain", nil
}

// ProcessArquivo lê o conteúdo do reader, calcula o hash, armazena no storage,
// extrai o conteúdo textual e persiste o arquivo no banco de dados. Retorna o arquivo
// existente caso o hash já exista.
func (w *DownloadProcessoWorker) ProcessArquivo(ctx context.Context, r io.Reader, contentType string) (*database.Arquivo, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler conteúdo do arquivo: %w", err)
	}

	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])

	arq, err := w.store.GetArquivo(ctx, hash)
	if err == nil {
		return arq, nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	storageKey := fmt.Sprintf("arquivos/%s", hash)

	err = w.storage.Put(ctx, storageKey, bytes.NewReader(body), contentType)
	if err != nil {
		return nil, fmt.Errorf("falha ao armazenar arquivo: %w", err)
	}

	conteudo, formato, err := w.extractContent(ctx, body, contentType)
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

	err = w.store.SaveArquivo(ctx, arq)
	if err != nil {
		return nil, fmt.Errorf("falha ao salvar arquivo: %w", err)
	}

	return arq, nil
}

func (w *DownloadProcessoWorker) Work(ctx context.Context, job *river.Job[DownloadProcessoArgs]) error {
	p, err := w.store.GetProcesso(ctx, job.Args.ProcessoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			return river.JobCancel(err)
		default:
			return fmt.Errorf("failed to get processo: %w", err)
		}
	}

	p.StatusProcessamento = "PROCESSANDO"
	err = w.store.UpdateProcesso(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to update processo: %w", err)
	}

	docs, err := w.sei.ListarDocumentos(ctx, p.LinkAcesso)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	g.SetLimit(3)

	var mu sync.Mutex
	dd := make([]*database.Documento, 0, len(docs))

	for _, doc := range docs {
		g.Go(func() error {
			resp, err := w.sei.ConsultarDocumento(ctx, doc.Numero)
			if err != nil {
				var soapErr *soap.Error
				switch {
				case errors.As(err, &soapErr):
					return nil
				default:
					return fmt.Errorf("failed to consultar documento: %w", err)
				}
			}

			b, err := json.Marshal(resp.Parametros)
			if err != nil {
				return fmt.Errorf("failed to marshal documento: %w", err)
			}

			res, err := http.Get(resp.Parametros.LinkAcesso)
			if err != nil {
				return fmt.Errorf("failed to request documento: %w", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to request documento: %d", res.StatusCode)
			}

			arq, err := w.ProcessArquivo(ctx, res.Body, res.Header.Get("Content-Type"))
			if err != nil {
				return fmt.Errorf("failed to process arquivo: %w", err)
			}

			tipo := resp.Parametros.Serie.Nome
			if resp.Parametros.Numero != "" {
				tipo += " " + resp.Parametros.Numero
			}

			mu.Lock()
			defer mu.Unlock()

			dd = append(dd, &database.Documento{
				Numero:       doc.Numero,
				ProcessoID:   p.ID,
				LinkAcesso:   resp.Parametros.LinkAcesso,
				Tipo:         tipo,
				Unidade:      resp.Parametros.UnidadeElaboradora.Sigla,
				MetadadosAPI: b,
				ArquivoHash:  arq.Hash,
			})
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to fetch documento: %w", err)
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback(ctx)

	store := w.store.WithTx(tx)
	for _, d := range dd {
		err = store.UpsertDocumento(ctx, d)
		if err != nil {
			return fmt.Errorf("failed to upsert documento: %w", err)
		}
	}

	p.StatusProcessamento = "SUCESSO"
	err = store.UpdateProcesso(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to update processo: %w", err)
	}

	client := river.ClientFromContext[pgx.Tx](ctx)
	_, err = client.InsertTx(ctx, tx, AnalisarProcessoArgs{
		ProcessoID: p.ID,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to insert analise task: %w", err)
	}

	return tx.Commit(ctx)
}

func (w *DownloadProcessoWorker) Timeout(job *river.Job[DownloadProcessoArgs]) time.Duration {
	return DownloadProcessoTimeout
}
