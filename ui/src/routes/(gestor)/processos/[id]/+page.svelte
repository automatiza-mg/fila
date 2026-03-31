<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import PrioridadeForm from "$lib/components/prioridade-form.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { calcularIdade } from "$lib/date";
  import { formatCpf } from "$lib/formatter";
  import { hasPapel } from "$lib/papel";
  import { statusText } from "$lib/processo";
  import ArrowSquareOutIcon from "phosphor-svelte/lib/ArrowSquareOutIcon";
  import type { PageProps } from "./$types";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";

  let { data }: PageProps = $props();
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

  <div class="flex items-center gap-4">
    <Button
      variant="outline"
      href="/preview/{data.processo.id}"
      target="_blank"
      rel="noopener noreferrer"
    >
      Visualizar PDF
      <ArrowSquareOutIcon />
    </Button>
    {#if !data.processo.prioridade && hasPapel(data.usuario, "GESTOR")}
      <PrioridadeForm paId={data.processo.id} />
    {/if}
  </div>
</div>
