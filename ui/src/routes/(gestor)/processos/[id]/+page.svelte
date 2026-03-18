<script lang="ts">
  import { statusText } from "$lib/processo";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processo {data.processo.numero} - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div>
    <span>Nº processo SEI:</span>
    <span>{data.processo.numero}</span>
  </div>

  <div class="flex gap-4 max-w-xl">
    <div class="flex-1 space-y-4">
      <div class="grid gap-1">
        <label for="analista">Analista Responsável</label>
        <input
          type="text"
          readonly
          id="analista"
          value={data.processo.analista ?? "Não possui"}
          class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        />
      </div>

      <div class="grid gap-1">
        <label for="status">Status</label>
        <input
          type="text"
          readonly
          id="status"
          value={statusText(data.processo.status)}
          class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        />
      </div>

      <div class="grid gap-1">
        <label for="prioridade">Prioritário</label>
        <input
          type="text"
          readonly
          id="prioridade"
          value={data.processo.prioridade ? "Sim" : "Não"}
          class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        />
        {#if !data.processo.prioridade}
          <div>
            <button class="text-smè">Solicitar Prioridade</button>
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 space-y-4">
      <div class="grid gap-1">
        <label for="data-requerimento">Data Requerimento</label>
        <input
          type="text"
          readonly
          id="data-requerimento"
          value={new Date(data.processo.data_requerimento).toLocaleDateString(
            "pt-BR",
            {
              timeZone: "UTC",
            },
          )}
          class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        />
      </div>

      <div class="grid gap-1">
        <label for="score">Score</label>
        <input
          type="text"
          readonly
          id="score"
          value={data.processo.score}
          class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
        />
      </div>
    </div>
  </div>
</div>
