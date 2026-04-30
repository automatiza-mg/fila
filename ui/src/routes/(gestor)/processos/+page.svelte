<script lang="ts">
  import ProcessoForm from "$lib/components/processo-form.svelte";
  import RecalcularScoresDialog from "$lib/components/recalcular-scores-dialog.svelte";
  import ServidorPopover from "$lib/components/servidor-popover.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import Pagination from "$lib/components/ui/pagination.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import { hasPapel } from "$lib/papel";
  import { statusColor, statusText } from "$lib/processo";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processos - Fila Aposentadoria</title>
</svelte:head>

<div class="flex grow flex-col gap-6">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
    <form method="GET" class="flex items-center gap-2">
      <Input
        name="numero"
        type="text"
        value={data.numero}
        placeholder="Número do processo"
      />
      <Select name="status" aria-label="Status" class="min-w-30" value={data.status}>
        <option value="">Todos</option>
        <option value="ANALISE_PENDENTE">Análise Pendente</option>
        <option value="EM_ANALISE">Em Análise</option>
        <option value="EM_DILIGENCIA">Em Diligência</option>
        <option value="RETORNO_DILIGENCIA">Retorno Diligência</option>
        <option value="LEITURA_INVALIDA">Leitura Inválida</option>
        <option value="CONCLUIDO">Concluído</option>
      </Select>
      <Button type="submit">Buscar</Button>
    </form>

    <div class="flex flex-wrap items-center gap-2">
      <!-- TODO: remover gate de ADMIN quando processos vierem do datalake -->
      {#if hasPapel(data.usuario, "ADMIN")}
        <ProcessoForm />
      {/if}
      <RecalcularScoresDialog />
      <Button href="/processos/prioridades">Solicitações de Prioridade</Button>
    </div>
  </div>

  <div>
    {#if data.processos.data.length === 0}
      <div class="p-6 border border-border-strong">
        <p class="text-center font-semibold">Nenhum processo encontrado.</p>
        <p class="text-center text-sm text-muted-foreground">
          Ajuste os filtros ou aguarde novos processos serem cadastrados.
        </p>
      </div>
    {:else}
      <table class="w-full border border-border text-sm">
        <thead>
          <tr class="border-y border-border bg-surface-alt">
            <th scope="col" class="text-left font-semibold p-2.5">Numero</th>
            <th scope="col" class="text-left font-semibold p-2.5"
              >Requerente</th
            >
            <th scope="col" class="text-left font-semibold p-2.5">Status</th>
            <th scope="col" class="text-left font-semibold p-2.5">
              Data Requerimento
            </th>
            <th scope="col" class="text-left font-semibold p-2.5">Score</th>
            <th scope="col" class="text-left font-semibold p-2.5"
              >Prioritário</th
            >
            <th scope="col" class="text-left font-semibold p-2.5">Analista</th>
          </tr>
        </thead>
        <tbody>
          {#each data.processos.data as processo (processo.id)}
            <tr class="border-b border-border hover:bg-surface-subtle">
              <td class="p-2.5">
                <a
                  class="text-primary underline"
                  href={`/processos/${processo.id}`}
                >
                  {processo.numero}</a
                >
              </td>
              <td class="p-2.5">
                <ServidorPopover cpf={processo.cpf_requerente} />
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
    {/if}
  </div>

  <Pagination data={data.processos} />
</div>
