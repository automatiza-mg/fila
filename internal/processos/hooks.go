package processos

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// AnalyzeHook é executado quando a análise de um processo é concluída.
// A implementação pode usar a transação passada para garantir consistência atômica.
type AnalyzeHook interface {
	OnAnalyzeCompleteTx(ctx context.Context, tx pgx.Tx, p *Processo, dd []*Documento) error
}

// RegisterHook registra um novo [AnalyzeHook] no serviço.
func (s *Service) RegisterHook(h AnalyzeHook) {
	s.hooks = append(s.hooks, h)
}

// notifyAnalyzeCompleteTx executa todos os hooks registrados dentro da mesma transação.
// Se um hook falhar, a transação será revertida pelo chamador.
func (s *Service) notifyAnalyzeCompleteTx(ctx context.Context, tx pgx.Tx, p *Processo, dd []*Documento) error {
	for _, h := range s.hooks {
		if err := h.OnAnalyzeCompleteTx(ctx, tx, p, dd); err != nil {
			return err
		}
	}
	return nil
}
