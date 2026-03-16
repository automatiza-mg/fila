package pipeline

import (
	"context"
	"fmt"
)

// Step representa uma etapa individual do pipeline de análise de processos SEI.
type Step interface {
	Name() string
	Run(ctx context.Context, state *State) error
}

// StepFunc adapta uma função para a interface [Step].
type StepFunc struct {
	name string
	fn   func(ctx context.Context, state *State) error
}

// Name retorna o nome do step.
func (s StepFunc) Name() string { return s.name }

// Run executa a função associada ao step.
func (s StepFunc) Run(ctx context.Context, state *State) error { return s.fn(ctx, state) }

// NewStep cria um [Step] a partir de uma função.
func NewStep(name string, fn func(ctx context.Context, state *State) error) Step {
	return StepFunc{name: name, fn: fn}
}

// Pipeline executa uma sequência de [Step]s em ordem.
// Se um step falha, a execução é interrompida imediatamente.
type Pipeline struct {
	steps []Step
}

// New cria um novo [Pipeline] com os steps fornecidos.
func New(steps ...Step) *Pipeline {
	return &Pipeline{steps: steps}
}

// Run executa todos os steps do pipeline em sequência.
func (p *Pipeline) Run(ctx context.Context, state *State) error {
	for _, s := range p.steps {
		if err := s.Run(ctx, state); err != nil {
			return fmt.Errorf("%s: %w", s.Name(), err)
		}
	}
	return nil
}
