<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

  function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString("pt-BR");
  }

  const statusLabels: Record<string, string> = {
    ANALISE_PENDENTE: "Análise Pendente",
    EM_ANALISE: "Em Análise",
    EM_DILIGENCIA: "Em Diligência",
    RETORNO_DILIGENCIA: "Retorno Diligência",
    LEITURA_INVALIDA: "Leitura Inválida",
    CONCLUIDO: "Concluído",
  };

  const statusColors: Record<string, string> = {
    ANALISE_PENDENTE: "bg-yellow-100 text-yellow-800",
    EM_ANALISE: "bg-blue-100 text-blue-800",
    EM_DILIGENCIA: "bg-orange-100 text-orange-800",
    RETORNO_DILIGENCIA: "bg-purple-100 text-purple-800",
    LEITURA_INVALIDA: "bg-red-100 text-red-800",
    CONCLUIDO: "bg-green-100 text-green-800",
  };
</script>

<svelte:head>
  <title>Processos | Fila Aposentadoria</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 p-6">
  <div class="max-w-7xl mx-auto">
    <div class="mb-6">
      <h1 class="text-3xl font-bold text-escritorio">
        Processos de Aposentadoria
      </h1>
      <p class="text-gray-600 mt-1">
        Total: {data.processos?.total_count ?? 0} processos
      </p>
    </div>

    {#if data.processos?.data && data.processos.data.length > 0}
      <div class="overflow-x-auto bg-white rounded-lg shadow">
        <table class="w-full">
          <thead>
            <tr class="border-b border-gray-200 bg-escritorio">
              <th class="px-6 py-3 text-left text-sm font-semibold text-white"
                >Número</th
              >
              <th class="px-6 py-3 text-left text-sm font-semibold text-white"
                >Status</th
              >
              <th class="px-6 py-3 text-left text-sm font-semibold text-white">
                Data Requerimento
              </th>
              <th class="px-6 py-3 text-left text-sm font-semibold text-white">
                CPF Requerente
              </th>
              <th class="px-6 py-3 text-center text-sm font-semibold text-white"
                >Prioridade</th
              >
              <th class="px-6 py-3 text-center text-sm font-semibold text-white"
                >Score</th
              >
            </tr>
          </thead>
          <tbody>
            {#each data.processos.data as processo}
              <tr
                class="border-b border-gray-100 hover:bg-gray-50 transition-colors"
              >
                <td class="px-6 py-4 text-sm font-medium text-escritorio">
                  <a href="/processos/{processo.id}"> {processo.numero}</a>
                </td>
                <td class="px-6 py-4 text-sm">
                  <div
                    class="inline-block px-3 py-1 rounded-full text-xs font-semibold whitespace-nowrap {statusColors[
                      processo.status
                    ] ?? 'bg-gray-100 text-gray-800'}"
                  >
                    {statusLabels[processo.status] ?? processo.status}
                  </div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-700">
                  {formatDate(processo.data_requerimento)}
                </td>
                <td class="px-6 py-4 text-sm text-gray-600 font-mono">
                  {processo.cpf_requerente}
                </td>
                <td class="px-6 py-4 text-center text-sm">
                  {#if processo.prioridade}
                    <span
                      class="inline-block px-2 py-0.5 rounded-full text-xs font-semibold bg-red-100 text-red-700"
                    >
                      Alta
                    </span>
                  {:else}
                    <span class="text-gray-400">—</span>
                  {/if}
                </td>
                <td class="px-6 py-4 text-center text-sm">
                  {processo.score}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if data.processos.total_pages > 1}
        <div
          class="mt-4 flex items-center justify-between text-sm text-gray-600"
        >
          <p>
            Página {data.processos.current_page} de {data.processos.total_pages}
            — Mostrando {(data.processos.current_page - 1) *
              data.processos.limit +
              1} a {Math.min(
              data.processos.current_page * data.processos.limit,
              data.processos.total_count,
            )} de {data.processos.total_count} processos
          </p>
        </div>
      {/if}
    {:else}
      <div class="bg-white rounded-lg shadow p-12 text-center">
        <p class="text-gray-500 text-lg">Nenhum processo encontrado.</p>
      </div>
    {/if}
  </div>
</div>
