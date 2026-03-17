<script lang="ts">
  import { entrarForm } from "../auth.remote";
</script>

<svelte:head>
  <title>Entrar - Fila Aposentadoria</title>
</svelte:head>

<h1 class="text-3xl font-bold text-center">Entrar</h1>

{#each entrarForm.fields.issues() as issue}
  <div role="alert" class="p-4 bg-red-50 border border-red-300 rounded-2xl">
    {issue.message}
  </div>
{/each}

<form class="flex flex-col gap-8" {...entrarForm}>
  <div class="flex flex-col gap-4">
    <div class="grid gap-1">
      <label for="cpf">CPF</label>
      <input
        {...entrarForm.fields.cpf.as("text")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
      />
      {#each entrarForm.fields.cpf.issues() as issue}
        <p class="text-sm text-red-600">{issue.message}</p>
      {/each}
    </div>

    <div class="grid gap-1">
      <label for="senha">Senha</label>
      <input
        {...entrarForm.fields._senha.as("password")}
        class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
      />
      <div class="flex justify-end">
        <a href="/recuperar-senha" class="text-muted-foreground underline"
          >Esqueci minha senha</a
        >
      </div>
      {#each entrarForm.fields._senha.issues() as issue}
        <p class="text-sm text-red-600">{issue.message}</p>
      {/each}
    </div>
  </div>

  <button
    class="px-4 py-2 font-semibold bg-primary text-white rounded-2xl border border-transparent"
  >
    Enviar
  </button>
</form>
