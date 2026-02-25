<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

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
</script>

<svelte:head>
  <title>Usuários | Fila Aposentadoria</title>
</svelte:head>

<div class="mb-6">
  <h1 class="text-3xl font-bold text-escritorio">Usuários</h1>
  <p class="text-gray-600 mt-1">
    Total: {data.usuarios?.length ?? 0} usuários
  </p>
</div>

{#if data.usuarios && data.usuarios.length > 0}
  <div class="overflow-x-auto bg-white rounded-lg shadow">
    <table class="w-full">
      <thead>
        <tr class="border-b border-gray-200 bg-escritorio">
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >Nome</th
          >
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >CPF</th
          >
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >Email</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Email Verificado</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Papel</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Pendências</th
          >
        </tr>
      </thead>
      <tbody>
        {#each data.usuarios as usuario}
          <tr
            class="border-b border-gray-100 hover:bg-gray-50 transition-colors"
          >
            <td class="px-6 py-4 text-sm font-medium text-gray-900">
              {usuario.nome}
            </td>
            <td class="px-6 py-4 text-sm text-gray-600 font-mono">
              {usuario.cpf}
            </td>
            <td class="px-6 py-4 text-sm text-gray-700">
              {usuario.email}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.email_verificado}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800"
                >
                  Sim
                </span>
              {:else}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-800"
                >
                  Não
                </span>
              {/if}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.papel}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold {papelColors[
                    usuario.papel
                  ] ?? 'bg-gray-100 text-gray-800'}"
                >
                  {papelLabels[usuario.papel] ?? usuario.papel}
                </span>
              {:else}
                <span class="text-gray-400">—</span>
              {/if}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.pendencias && usuario.pendencias.length > 0}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-yellow-100 text-yellow-800"
                >
                  {usuario.pendencias.length}
                </span>
              {:else}
                <span class="text-gray-400">—</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <div class="bg-white rounded-lg shadow p-12 text-center">
    <p class="text-gray-500 text-lg">Nenhum usuário encontrado.</p>
  </div>
{/if}
