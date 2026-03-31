<script lang="ts">
  import Button from "$lib/components/ui/button.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { statusColor, statusText } from "$lib/processo";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processos - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <form method="GET" class="flex items-center gap-2">
      <Input
        name="numero"
        type="text"
        value={data.numero}
        placeholder="Número do processo"
        required
      />
      <Button type="submit">Buscar</Button>
    </form>

    <Button href="/processos/prioridades">Solicitações de Prioridade</Button>
  </div>

  <div>
    <table class="w-full border border-border text-sm">
      <thead>
        <tr class="border-y border-border bg-surface-alt">
          <th scope="col" class="text-left font-semibold p-2.5">Numero</th>
          <th scope="col" class="text-left font-semibold p-2.5">Status</th>
          <th scope="col" class="text-left font-semibold p-2.5">
            Data Requerimento
          </th>
          <th scope="col" class="text-left font-semibold p-2.5">Score</th>
          <th scope="col" class="text-left font-semibold p-2.5">Prioritário</th>
          <th scope="col" class="text-left font-semibold p-2.5">Analista</th>
        </tr>
      </thead>
      <tbody>
        {#each data.processos.data as processo}
          <tr class="border-b border-border hover:bg-surface-subtle">
            <td class="p-2.5">
              <a
                class="text-primary underline"
                href={`/processos/${processo.id}`}
              >
                {processo.numero}</a
              >
            </td>
            <td class={`p-2.5 ${statusColor(processo.status)}`}
              >{statusText(processo.status)}</td
            >
            <td class="p-2.5"
              >{new Date(processo.data_requerimento).toLocaleDateString(
                "pt-BR",
                {
                  timeZone: "UTC",
                },
              )}</td
            >
            <td class="p-2.5">{processo.score}</td>
            <td class="p-2.5">{processo.prioridade ? "Sim" : "Não"}</td>
            <td class="p-2.5">
              {#if processo.analista}
                <a
                  class="text-primary underline"
                  href="/usuarios/{processo.analista_id}"
                >
                  {processo.analista}
                </a>
              {:else}
                Não possui
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
