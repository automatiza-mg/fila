package processos

import "context"

// AnalyzeHook é executado quando a análise de um processo é concluída.
type AnalyzeHook interface {
	OnAnalyzeComplete(ctx context.Context, p *Processo, dd []*Documento) error
}

// RegisterHook registra um novo [AnalyzeHook] no serviço.
func (s *Service) RegisterHook(h AnalyzeHook) {
	s.hooks = append(s.hooks, h)
}

// notifyAnalyzeComplete executa todos os hooks registrados.
func (s *Service) notifyAnalyzeComplete(ctx context.Context, p *Processo, dd []*Documento) error {
	for _, h := range s.hooks {
		if err := h.OnAnalyzeComplete(ctx, p, dd); err != nil {
			return err
		}
	}
	return nil
}
