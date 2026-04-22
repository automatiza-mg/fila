<script lang="ts">
  import ServidorPopover from "$lib/components/servidor-popover.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import Pagination from "$lib/components/ui/pagination.svelte";
  import { statusColor, statusText } from "$lib/processo";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Histórico de Processos - Fila Aposentadoria</title>
</svelte:head>

<div class="flex grow flex-col gap-6">
  <form method="GET" class="flex items-center gap-2">
    <Input
      name="numero"
      type="text"
      value={data.numero}
      placeholder="Número do processo"
    />
    <Button type="submit">Buscar</Button>
  </form>

  {#if data.processos.data.length === 0}
    <div class="flex flex-col grow items-center justify-center">
      <p class="font-semibold">Nenhum processo no histórico.</p>
      <p class="text-muted-foreground text-xs">
        Os processos concluídos, marcados como leitura inválida ou enviados para
        diligência aparecerão aqui.
      </p>
    </div>
  {:else}
    <div>
      <table class="w-full border border-border text-sm">
        <thead>
          <tr class="border-y border-border bg-surface-alt">
            <th scope="col" class="text-left font-semibold p-2.5">Número</th>
            <th scope="col" class="text-left font-semibold p-2.5">Requerente</th>
            <th scope="col" class="text-left font-semibold p-2.5">Status</th>
            <th scope="col" class="text-left font-semibold p-2.5">
              Data Requerimento
            </th>
            <th scope="col" class="text-left font-semibold p-2.5">
              Atualizado em
            </th>
          </tr>
        </thead>
        <tbody>
          {#each data.processos.data as processo}
            <tr class="border-b border-border hover:bg-surface-subtle">
              <td class="p-2.5">{processo.numero}</td>
              <td class="p-2.5">
                <ServidorPopover cpf={processo.cpf_requerente} />
              </td>
              <td class={`p-2.5 ${statusColor(processo.status)}`}>
                {statusText(processo.status)}
              </td>
              <td class="p-2.5">
                {new Date(processo.data_requerimento).toLocaleDateString(
                  "pt-BR",
                  { timeZone: "UTC" },
                )}
              </td>
              <td class="p-2.5">
                {new Date(processo.atualizado_em).toLocaleDateString("pt-BR")}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <Pagination data={data.processos} />
  {/if}
</div>
