package pipeline

import (
	"context"
)

var (
	_ DocumentLister    = (*fakeLister)(nil)
	_ DocumentFetcher   = (*fakeFetcher)(nil)
	_ DocumentChecker   = (*fakeChecker)(nil)
	_ StatusUpdater     = (*fakeStatusUpdater)(nil)
	_ DocumentPersister = (*fakePersister)(nil)
)

type fakeLister struct {
	docs []string
	err  error
}

func (f *fakeLister) ListDocumentos(_ context.Context, _ string) ([]string, error) {
	return f.docs, f.err
}

type fakeFetcher struct {
	docs []DocBuscado
	err  error
}

func (f *fakeFetcher) FetchDocumentos(_ context.Context, _ []string) ([]DocBuscado, error) {
	return f.docs, f.err
}

type fakeChecker struct {
	existing map[string]bool
	err      error
}

func (f *fakeChecker) ExisteDocumento(_ context.Context, numero string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.existing[numero], nil
}

type fakeStatusUpdater struct {
	called bool
	status string
	err    error
}

func (f *fakeStatusUpdater) AtualizarStatus(_ context.Context, state *State) error {
	f.called = true
	f.status = state.Status
	return f.err
}

type fakePersister struct {
	called bool
	docs   []DocBuscado
	err    error
}

func (f *fakePersister) PersistirDocumentos(_ context.Context, state *State) error {
	f.called = true
	f.docs = state.DocumentosBuscados
	return f.err
}
