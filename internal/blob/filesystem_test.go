package blob

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestFilesystem_Lifecycle(t *testing.T) {
	t.Parallel()

	storage, err := NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	r := strings.NewReader("Hello World!")

	err = storage.Put(t.Context(), "file.txt", r, "text/plain")
	if err != nil {
		t.Fatal(err)
	}

	rc, err := storage.Get(t.Context(), "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "Hello World!" {
		t.Fatal("unexpected content found")
	}

	_, err = storage.Get(t.Context(), "bogus.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	err = storage.Delete(t.Context(), "file.txt")
	if err != nil {
		t.Fatal(err)
	}
}
