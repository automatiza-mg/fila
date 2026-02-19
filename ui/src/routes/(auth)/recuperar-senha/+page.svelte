<svelte:head>
  <title>Esqueci minha senha | Fila Aposentadoria</title>
</svelte:head>

<div>
  <h1 class="text-center text-3xl font-bold">Esqueci minha senha</h1>
</div>

<!-- TODO: Adicionar formulário -->
<div>
  <div class="grid gap-1">
    <label for="cpf">CPF</label>
    <input type="text" name="cpf" id="cpf" />
  </div>

  <button>Enviar</button>
</div>

<div class="flex flex-col gap-1">
  <a href="/entrar" class="text-escritorio underline text-center">
    Lembrei minha senha
  </a>
</div>
