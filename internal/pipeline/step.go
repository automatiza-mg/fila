package pipeline

import (
	"context"
)

// DocumentLister lista os números dos documentos de um processo pela página de acesso do SEI.
type DocumentLister interface {
	ListDocumentos(ctx context.Context, linkAcesso string) ([]string, error)
}

// DocumentFetcher busca os documentos no SEI e extrai o texto via OCR.
type DocumentFetcher interface {
	FetchDocumentos(ctx context.Context, nums []string) ([]DocBuscado, error)
}

// DocumentChecker verifica se um documento já existe no banco de dados.
type DocumentChecker interface {
	ExisteDocumento(ctx context.Context, numero string) (bool, error)
}

// StatusUpdater atualiza o status de processamento de um processo.
type StatusUpdater interface {
	AtualizarStatus(ctx context.Context, state *State) error
}

// DocumentPersister salva documentos no banco de dados.
type DocumentPersister interface {
	PersistirDocumentos(ctx context.Context, state *State) error
}

// ListDocumentos cria um step que lista os documentos do processo via scraping do SEI.
func ListDocumentos(lister DocumentLister) Step {
	return NewStep("listar_documentos", func(ctx context.Context, state *State) error {
		docs, err := lister.ListDocumentos(ctx, state.LinkAcesso)
		if err != nil {
			return err
		}
		state.DocumentosListados = docs
		return nil
	})
}

// FiltrarNovos cria um step que filtra os documentos já existentes no banco de dados,
// mantendo apenas os que ainda precisam ser buscados.
func FiltrarNovos(checker DocumentChecker) Step {
	return NewStep("filtrar_novos", func(ctx context.Context, state *State) error {
		var novos []string
		for _, num := range state.DocumentosListados {
			exists, err := checker.ExisteDocumento(ctx, num)
			if err != nil {
				return err
			}
			if !exists {
				novos = append(novos, num)
			}
		}
		state.DocumentosListados = novos
		return nil
	})
}

// BuscarDocumentos cria um step que busca o conteúdo completo dos documentos novos
// no SEI (API + download + OCR).
func BuscarDocumentos(fetcher DocumentFetcher) Step {
	return NewStep("buscar_documentos", func(ctx context.Context, state *State) error {
		if len(state.DocumentosListados) == 0 {
			state.DocumentosBuscados = []DocBuscado{}
			return nil
		}

		docs, err := fetcher.FetchDocumentos(ctx, state.DocumentosListados)
		if err != nil {
			return err
		}
		state.DocumentosBuscados = docs
		return nil
	})
}

// AtualizarStatus cria um step que atualiza o status de processamento do processo.
func AtualizarStatus(status string, updater StatusUpdater) Step {
	return NewStep("atualizar_status", func(ctx context.Context, state *State) error {
		state.Status = status
		return updater.AtualizarStatus(ctx, state)
	})
}

// PersistirDocumentos cria um step que salva os documentos buscados no banco de dados.
func PersistirDocumentos(persister DocumentPersister) Step {
	return NewStep("persistir_documentos", func(ctx context.Context, state *State) error {
		return persister.PersistirDocumentos(ctx, state)
	})
}
