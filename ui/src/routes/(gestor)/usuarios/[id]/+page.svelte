<script lang="ts">
  import type { PageProps } from "./$types";
  import { enviarCadastro } from "./enviar-cadastro.remote";
  import { excluir } from "./excluir.remote";

  let { data }: PageProps = $props();

  let confirmDeleteOpen = $state(false);
  let confirmDialogEl: HTMLDialogElement | undefined = $state();

  const papelLabels: Record<string, string> = {
    ADMIN: "Administrador",
    ANALISTA: "Analista",
    GESTOR: "Gestor",
    SUBSECRETARIO: "Subsecretário",
  };

  const papelColors: Record<string, string> = {
    ADMIN: "bg-red-100 text-red-800",
    ANALISTA: "bg-blue-100 text-blue-800",
    GESTOR: "bg-purple-100 text-purple-800",
    SUBSECRETARIO: "bg-amber-100 text-amber-800",
  };

  function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString("pt-BR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function openConfirmDelete() {
    confirmDeleteOpen = true;
    confirmDialogEl?.showModal();
  }

  function closeConfirmDelete() {
    confirmDeleteOpen = false;
    confirmDialogEl?.close();
  }
</script>

<svelte:head>
  <title>{data.usuario.nome} | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div>
    <a
      href="/usuarios"
      class="text-sm text-escritorio-light hover:underline"
    >
      &larr; Voltar para Usuários
    </a>
  </div>

  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-3xl font-bold text-escritorio">{data.usuario.nome}</h1>
      <p class="text-gray-500 font-mono mt-1">{data.usuario.cpf}</p>
    </div>
    {#if data.usuario.papel}
      <span
        class="inline-block px-4 py-1.5 rounded-full text-sm font-semibold {papelColors[data.usuario.papel] ?? 'bg-gray-100 text-gray-800'}"
      >
        {papelLabels[data.usuario.papel] ?? data.usuario.papel}
      </span>
    {/if}
  </div>

  {#each enviarCadastro.fields.issues() as issue}
    <div class="p-3 bg-red-50 text-red-700 rounded-xl text-sm">
      {issue.message}
    </div>
  {/each}
  {#each excluir.fields.issues() as issue}
    <div class="p-3 bg-red-50 text-red-700 rounded-xl text-sm">
      {issue.message}
    </div>
  {/each}

  <div class="bg-white rounded-lg shadow overflow-hidden">
    <div class="px-6 py-4 bg-escritorio">
      <h2 class="text-lg font-semibold text-white">Informações</h2>
    </div>
    <div class="p-6">
      <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
        <div>
          <dt class="text-sm font-medium text-gray-500">Email</dt>
          <dd class="mt-1 text-sm text-gray-900">{data.usuario.email}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Email Verificado</dt>
          <dd class="mt-1">
            {#if data.usuario.email_verificado}
              <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800">
                Sim
              </span>
            {:else}
              <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-800">
                Não
              </span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Papel</dt>
          <dd class="mt-1 text-sm text-gray-900">
            {#if data.usuario.papel}
              {papelLabels[data.usuario.papel] ?? data.usuario.papel}
            {:else}
              <span class="text-gray-400">Não definido</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Pendências</dt>
          <dd class="mt-1 text-sm text-gray-900">
            {#if data.usuario.pendencias && data.usuario.pendencias.length > 0}
              <ul class="space-y-1">
                {#each data.usuario.pendencias as pendencia}
                  <li>
                    <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-yellow-100 text-yellow-800">
                      {pendencia.titulo}
                    </span>
                  </li>
                {/each}
              </ul>
            {:else}
              <span class="text-gray-400">Nenhuma</span>
            {/if}
          </dd>
        </div>
      </dl>
    </div>
  </div>

  {#if data.analista}
    <div class="bg-white rounded-lg shadow overflow-hidden">
      <div class="px-6 py-4 bg-escritorio">
        <h2 class="text-lg font-semibold text-white">Dados de Analista</h2>
      </div>
      <div class="p-6">
        <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
          <div>
            <dt class="text-sm font-medium text-gray-500">Órgão</dt>
            <dd class="mt-1 text-sm text-gray-900">{data.analista.orgao}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">Unidade</dt>
            <dd class="mt-1 text-sm text-gray-900">{data.analista.sei_unidade_sigla}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">Afastado</dt>
            <dd class="mt-1">
              {#if data.analista.afastado}
                <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-800">
                  Sim
                </span>
              {:else}
                <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800">
                  Não
                </span>
              {/if}
            </dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">Última Atribuição</dt>
            <dd class="mt-1 text-sm text-gray-900">
              {#if data.analista.ultima_atribuicao_em}
                {formatDate(data.analista.ultima_atribuicao_em)}
              {:else}
                <span class="text-gray-400">Nenhuma</span>
              {/if}
            </dd>
          </div>
        </dl>
      </div>
    </div>
  {/if}

  <div class="flex gap-3">
    {#if !data.usuario.email_verificado}
      <form {...enviarCadastro}>
        <input type="hidden" name="usuario_id" value={data.usuario.id} />
        <button
          type="submit"
          class="px-4 py-2 bg-escritorio-light text-white rounded-xl font-semibold hover:opacity-90 transition-opacity"
        >
          Reenviar Cadastro
        </button>
      </form>
    {/if}
    <button
      type="button"
      onclick={openConfirmDelete}
      class="px-4 py-2 bg-red-600 text-white rounded-xl font-semibold hover:opacity-90 transition-opacity"
    >
      Excluir Usuário
    </button>
  </div>
</div>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<dialog
  bind:this={confirmDialogEl}
  class="rounded-2xl shadow-xl p-0 w-full max-w-md backdrop:bg-black/50"
  onclose={closeConfirmDelete}
  onkeydown={(e) => e.key === "Escape" && closeConfirmDelete()}
>
  <div class="p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-2">Confirmar Exclusão</h2>
    <p class="text-gray-600 mb-6">
      Tem certeza que deseja excluir o usuário <strong>{data.usuario.nome}</strong>? Esta ação não pode ser desfeita.
    </p>
    <div class="flex gap-3">
      <button
        type="button"
        onclick={closeConfirmDelete}
        class="flex-1 px-4 py-2 border border-stone-300 text-stone-700 rounded-xl font-semibold hover:bg-stone-50 transition-colors"
      >
        Cancelar
      </button>
      <form {...excluir} class="flex-1">
        <input type="hidden" name="usuario_id" value={data.usuario.id} />
        <button
          type="submit"
          class="w-full px-4 py-2 bg-red-600 text-white rounded-xl font-semibold hover:opacity-90 transition-opacity"
        >
          Excluir
        </button>
      </form>
    </div>
  </div>
</dialog>
