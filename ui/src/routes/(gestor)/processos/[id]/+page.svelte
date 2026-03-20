<script lang="ts">
  import PrioridadeForm from "$lib/components/prioridade-form.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { hasPapel } from "$lib/papel";
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

      <div>
        <FormField label="Prioritário" id="prioridade">
          <Input
            type="text"
            readonly
            id="prioridade"
            value={data.processo.prioridade ? "Sim" : "Não"}
          />
        </FormField>
        {#if !data.processo.prioridade && hasPapel(data.usuario, "GESTOR")}
          <div>
            <PrioridadeForm paId={data.processo.id} />
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 space-y-4">
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

      <FormField label="Score" id="score">
        <Input type="text" readonly id="score" value={data.processo.score} />
      </FormField>
    </div>
  </div>
</div>
