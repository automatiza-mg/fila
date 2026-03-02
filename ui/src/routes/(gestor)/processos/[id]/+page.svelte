<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

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

  function formatDateShort(dateString: string): string {
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
  <title>{data.processo.numero} | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div>
    <a
      href="/processos"
      class="text-sm text-escritorio-light hover:underline"
    >
      &larr; Voltar para Processos
    </a>
  </div>

  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-3xl font-bold text-escritorio">{data.processo.numero}</h1>
      <p class="text-gray-500 mt-1">Processo de Aposentadoria #{data.processo.id}</p>
    </div>
    <span
      class="inline-block px-4 py-1.5 rounded-full text-sm font-semibold {statusColors[data.processo.status] ?? 'bg-gray-100 text-gray-800'}"
    >
      {statusLabels[data.processo.status] ?? data.processo.status}
    </span>
  </div>

  <div class="bg-white rounded-lg shadow overflow-hidden">
    <div class="px-6 py-4 bg-escritorio">
      <h2 class="text-lg font-semibold text-white">Dados da Aposentadoria</h2>
    </div>
    <div class="p-6">
      <dl class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-6 gap-y-4">
        <div>
          <dt class="text-sm font-medium text-gray-500">Data Requerimento</dt>
          <dd class="mt-1 text-sm text-gray-900">
            {formatDateShort(data.processo.data_requerimento)}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">CPF Requerente</dt>
          <dd class="mt-1 text-sm text-gray-900 font-mono">
            {data.processo.cpf_requerente}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Score</dt>
          <dd class="mt-1 text-sm text-gray-900 font-semibold">
            {data.processo.score}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Prioridade</dt>
          <dd class="mt-1">
            {#if data.processo.prioridade}
              <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-700">
                Alta
              </span>
            {:else}
              <span class="text-gray-400">Normal</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Judicial</dt>
          <dd class="mt-1">
            {#if data.processo.judicial}
              <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-orange-100 text-orange-800">
                Sim
              </span>
            {:else}
              <span class="text-gray-400">Não</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Invalidez</dt>
          <dd class="mt-1">
            {#if data.processo.invalidez}
              <span class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-orange-100 text-orange-800">
                Sim
              </span>
            {:else}
              <span class="text-gray-400">Não</span>
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Criado em</dt>
          <dd class="mt-1 text-sm text-gray-900">
            {formatDate(data.processo.criado_em)}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Atualizado em</dt>
          <dd class="mt-1 text-sm text-gray-900">
            {formatDate(data.processo.atualizado_em)}
          </dd>
        </div>
      </dl>
    </div>
  </div>

  {#if data.processoBase}
    <div class="bg-white rounded-lg shadow overflow-hidden">
      <div class="px-6 py-4 bg-escritorio">
        <h2 class="text-lg font-semibold text-white">Processo SEI</h2>
      </div>
      <div class="p-6">
        <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
          <div>
            <dt class="text-sm font-medium text-gray-500">Unidade</dt>
            <dd class="mt-1 text-sm text-gray-900">
              {data.processoBase.sei_unidade_sigla}
            </dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">Link de Acesso</dt>
            <dd class="mt-1 text-sm">
              <a
                href={data.processoBase.link_acesso}
                target="_blank"
                rel="noopener noreferrer"
                class="text-escritorio-light hover:underline"
              >
                Abrir no SEI
              </a>
            </dd>
          </div>
          {#if data.processoBase.analisado_em}
            <div>
              <dt class="text-sm font-medium text-gray-500">Analisado em</dt>
              <dd class="mt-1 text-sm text-gray-900">
                {formatDate(data.processoBase.analisado_em)}
              </dd>
            </div>
          {/if}
        </dl>
      </div>
    </div>
  {/if}

  {#if data.documentos.length > 0}
    <div class="bg-white rounded-lg shadow overflow-hidden">
      <div class="px-6 py-4 bg-escritorio">
        <h2 class="text-lg font-semibold text-white">
          Documentos ({data.documentos.length})
        </h2>
      </div>
      <div class="divide-y divide-gray-100">
        {#each data.documentos as documento}
          <div class="px-6 py-3 flex items-center justify-between">
            <span class="text-sm text-gray-900">{documento.nome}</span>
            <a
              href={documento.link}
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-escritorio-light hover:underline"
            >
              Abrir
            </a>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if data.historico.length > 0}
    <div class="bg-white rounded-lg shadow overflow-hidden">
      <div class="px-6 py-4 bg-escritorio">
        <h2 class="text-lg font-semibold text-white">Histórico de Status</h2>
      </div>
      <div class="p-6">
        <div class="space-y-4">
          {#each data.historico as evento}
            <div class="flex gap-4 items-start">
              <div class="flex-shrink-0 w-2 h-2 mt-2 rounded-full bg-escritorio-light"></div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <span
                    class="inline-block px-2.5 py-0.5 rounded-full text-xs font-semibold {statusColors[evento.status_anterior] ?? 'bg-gray-100 text-gray-800'}"
                  >
                    {statusLabels[evento.status_anterior] ?? evento.status_anterior}
                  </span>
                  <span class="text-gray-400">&rarr;</span>
                  <span
                    class="inline-block px-2.5 py-0.5 rounded-full text-xs font-semibold {statusColors[evento.status_novo] ?? 'bg-gray-100 text-gray-800'}"
                  >
                    {statusLabels[evento.status_novo] ?? evento.status_novo}
                  </span>
                </div>
                {#if evento.observacao}
                  <p class="text-sm text-gray-600 mt-1">{evento.observacao}</p>
                {/if}
                <p class="text-xs text-gray-400 mt-1">
                  {formatDate(evento.criado_em)}
                </p>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>
