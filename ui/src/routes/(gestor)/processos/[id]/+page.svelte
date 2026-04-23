<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import PrioridadeForm from "$lib/components/prioridade-form.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { calcularIdade } from "$lib/date";
  import { atualizarProcessoPreviewCmd } from "$lib/fns/processos.remote";
  import { formatCpf } from "$lib/formatter";
  import { hasPapel } from "$lib/papel";
  import { statusText, statusColor } from "$lib/processo";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import ArrowRightIcon from "phosphor-svelte/lib/ArrowRightIcon";
  import ArrowSquareOutIcon from "phosphor-svelte/lib/ArrowSquareOutIcon";
  import CalendarIcon from "phosphor-svelte/lib/CalendarIcon";
  import FilePdfIcon from "phosphor-svelte/lib/FilePdfIcon";
  import WarningIcon from "phosphor-svelte/lib/WarningIcon";
  import { toast } from "svelte-sonner";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
  let atualizandoPreview = $state(false);

  async function atualizarPreview() {
    if (atualizandoPreview) return;
    atualizandoPreview = true;
    try {
      await atualizarProcessoPreviewCmd({
        processoId: data.processo.processo_id,
      });
      toast.success(
        "Preview em atualização. Recarregue a página em alguns instantes.",
      );
    } catch {
      toast.error("Não foi possível atualizar o preview");
    } finally {
      atualizandoPreview = false;
    }
  }
</script>

<svelte:head>
  <title>Processo {data.processo.numero} - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex justify-between items-center">
    <NumeroProcesso numero={data.processo.numero} />

    <a href="/processos" class="flex flex-col items-center">
      <ArrowElbowUpLeftIcon class="size-5" />
      <span>Voltar</span>
    </a>
  </div>

  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 max-w-4xl">
    <FormField label="CPF Requerente" id="cpf">
      <Input
        type="text"
        readonly
        id="cpf"
        value={formatCpf(data.processo.cpf_requerente)}
      />
    </FormField>

    <FormField label="Data de Nascimento" id="nascimento">
      <Input
        type="text"
        readonly
        id="nascimento"
        value="{new Date(
          data.processo.data_nascimento_requerente,
        ).toLocaleDateString('pt-BR', {
          timeZone: 'UTC',
        })} ({calcularIdade(data.processo.data_nascimento_requerente)} anos)"
      />
    </FormField>

    <FormField label="Data Requerimento" id="data-requerimento">
      <Input
        type="text"
        readonly
        id="data-requerimento"
        value={new Date(data.processo.data_requerimento).toLocaleDateString(
          "pt-BR",
          { timeZone: "UTC" },
        )}
      />
    </FormField>

    <FormField label="Analista Responsável" id="analista">
      <Input
        type="text"
        readonly
        id="analista"
        value={data.processo.analista ?? "Não possui"}
      />
    </FormField>

    <FormField label="Status" id="status">
      <Input
        type="text"
        readonly
        id="status"
        value={statusText(data.processo.status)}
      />
    </FormField>

    <FormField label="Score" id="score">
      <Input type="text" readonly id="score" value={data.processo.score} />
    </FormField>

    <FormField label="Judicial" id="judicial">
      <Input
        type="text"
        readonly
        id="judicial"
        value={data.processo.judicial ? "Sim" : "Não"}
      />
    </FormField>

    <FormField label="Invalidez" id="invalidez">
      <Input
        type="text"
        readonly
        id="invalidez"
        value={data.processo.invalidez ? "Sim" : "Não"}
      />
    </FormField>

    <FormField label="Prioritário" id="prioridade">
      <Input
        type="text"
        readonly
        id="prioridade"
        value={data.processo.prioridade ? "Sim" : "Não"}
      />
    </FormField>
  </div>

  {#if data.processo.alertas && data.processo.alertas.length > 0}
    <div
      class="rounded-xl border border-amber-300 bg-amber-50 text-amber-900 px-4 py-3 text-sm max-w-4xl"
    >
      <p class="flex items-center gap-1 font-medium">
        <WarningIcon class="size-4" />
        Alertas
      </p>
      <ul class="mt-1 list-disc pl-5 space-y-0.5">
        {#each data.processo.alertas as alerta}
          <li>{alerta}</li>
        {/each}
      </ul>
    </div>
  {/if}

  <div class="flex items-center gap-4">
    {#if !data.processo.prioridade && hasPapel(data.usuario, "GESTOR")}
      <PrioridadeForm paId={data.processo.id} />
    {/if}
    {#if data.processo.possui_preview}
      <Button
        variant="outline"
        href="/preview/{data.processo.id}"
        target="_blank"
        rel="noopener noreferrer"
      >
        Visualizar PDF
        <ArrowSquareOutIcon />
      </Button>
    {/if}
    <Button
      variant="outline"
      disabled={atualizandoPreview}
      onclick={atualizarPreview}
    >
      <FilePdfIcon />
      Atualizar Preview
    </Button>
  </div>

  <!-- Histórico -->
  {#if data.historico.length > 0}
    <div class="space-y-4">
      <h2 class="text-base font-semibold">Histórico</h2>

      <div class="divide-y divide-border">
        {#each data.historico as entry}
          <div class="py-3 space-y-1">
            <div class="flex items-center gap-3 text-sm">
              <span
                class="text-muted-foreground text-xs shrink-0 flex items-center gap-1"
              >
                <CalendarIcon class="size-4" />
                {new Date(entry.alterado_em).toLocaleString("pt-BR")}
              </span>

              <div class="flex items-center gap-1.5">
                {#if entry.status_anterior}
                  <span
                    class="inline-block rounded-md px-2 py-1 text-xs font-medium {statusColor(
                      entry.status_anterior,
                    )}"
                  >
                    {statusText(entry.status_anterior)}
                  </span>
                  <ArrowRightIcon class="size-4 text-muted-foreground" />
                {/if}
                <span
                  class="inline-block rounded-md px-2 py-1 text-xs font-medium {statusColor(
                    entry.status_novo,
                  )}"
                >
                  {statusText(entry.status_novo)}
                </span>
              </div>
            </div>

            {#if entry.observacao}
              <p class="text-xs p-0.5 border-l-2 border-border-strong pl-2">
                {entry.observacao}
              </p>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
