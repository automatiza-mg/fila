package auth

import (
	"testing"
)

func TestUsuarioLifecycle(t *testing.T) {
	t.Parallel()

	auth := newTestService(t)
	queue := auth.queue.(*fakeTaskInserter)

	counter := &fakeCounterHook{}
	auth.RegisterHook(counter)

	// Cria um novo usuário
	var setupToken string
	u, err := auth.CreateUsuario(t.Context(), CreateUsuarioParams{
		Nome:  "Fulano da Silva",
		CPF:   "123.456.789-09",
		Email: "fulano@email.com",
		Papel: PapelAnalista,
		TokenURL: func(token string) string {
			setupToken = token
			return "/cadatro/" + token
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if counter.actions != 1 {
		t.Fatal("CreateUsuario missed GetActions call")
	}

	// Verifica se a task foi inserida.
	if len(queue.Args()) != 1 {
		t.Fatal("expected a task to be inserted")
	}

	// Conclui o cadastro do usuário
	err = auth.SetupUsuario(t.Context(), SetupUsuarioParams{
		Token: setupToken,
		Senha: "xyyz",
	})
	if err != nil {
		t.Fatal(err)
	}

	u, err = auth.GetUsuario(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !u.HasSenha() || !u.EmailVerificado {
		t.Fatal("expected user setup to be done")
	}
	if counter.actions != 2 {
		t.Fatal("GetUsuario missed GetActions call")
	}

	// Atualizamos o papel para verificar se Cleanup é chamada.
	err = auth.UpdateUsuario(t.Context(), UpdateUsuarioParams{
		UsuarioID: u.ID,
		Nome:      u.Nome,
		Papel:     PapelGestor,
	})
	if err != nil {
		t.Fatal(err)
	}
	if counter.cleanups != 1 {
		t.Fatal("UpdateUsuario missed Cleanup call")
	}

	// Ao deletar o usuário, Cleanup é sempre chamada.
	err = auth.DeleteUsuario(t.Context(), u)
	if err != nil {
		t.Fatal(err)
	}
	if counter.cleanups != 2 {
		t.Fatal("DeleteUsuario missed Cleanup call")
	}

	_, err = auth.GetUsuario(t.Context(), u.ID)
	if err == nil {
		t.Fatal("expected error")
	}
}
