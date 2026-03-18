<script lang="ts">
  import Alert from "$lib/components/ui/alert.svelte";
  import { formatCpf } from "$lib/formatter";
  import { createUsuarioForm } from "../usuario.remote";

  $effect(() => {
    const value = createUsuarioForm.fields.cpf.value() ?? "";
    createUsuarioForm.fields.cpf.set(formatCpf(value));
  });
</script>

<svelte:head>
  <title>Cadastrar Usuário | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <h1 class="text-2xl font-semibold text-center">Cadastrar Usuário</h1>

  <form {...createUsuarioForm} class="flex flex-col gap-4 max-w-md mx-auto">
    {#each createUsuarioForm.fields.issues() as issue}
      <Alert message={issue.message} variant="danger" />
    {/each}

    <div class="grid gap-1">
      <label for="nome">Nome</label>
      <input
        id="nome"
        {...createUsuarioForm.fields.nome.as("text")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        required
      />
    </div>

    <div class="grid gap-1">
      <label for="cpf">CPF</label>
      <input
        id="cpf"
        {...createUsuarioForm.fields.cpf.as("text")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        required
      />

      {#each createUsuarioForm.fields.cpf.issues() as issue}
        <p class="text-sm text-red-500">{issue.message}</p>
      {/each}
    </div>

    <div class="grid gap-1">
      <label for="email">Email</label>
      <input
        id="email"
        {...createUsuarioForm.fields.email.as("email")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        required
      />

      {#each createUsuarioForm.fields.email.issues() as issue}
        <p class="text-sm text-red-500">{issue.message}</p>
      {/each}
    </div>

    <div class="grid gap-1">
      <label for="papel">Papel</label>
      <select
        id="papel"
        {...createUsuarioForm.fields.papel.as("select")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        required
      >
        <option value="ANALISTA">Analista</option>
        <option value="GESTOR">Gestor(a)</option>
        <option value="SUBSECRETARIO">Subsecretário(a)</option>
      </select>

      {#each createUsuarioForm.fields.papel.issues() as issue}
        <p class="text-sm text-red-500">{issue.message}</p>
      {/each}
    </div>

    <button
      class="px-4 py-2 font-semibold bg-primary text-white rounded-2xl border border-transparent"
    >
      Enviar
    </button>
  </form>
</div>
