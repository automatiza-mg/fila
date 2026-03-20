<script lang="ts">
  import FormField from "$lib/components/ui/form-field.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>{data.usuario.nome} - Fila Aposentadoria</title>
</svelte:head>

<div>
  Nome: {data.usuario.nome}
</div>

{#if data.usuario.papel === "ANALISTA"}
  {#if data.analista}
    Dados Analista
  {:else}
    <div class="flex flex-col gap-4 max-w-sm">
      <FormField label="Órgao de Exercício" id="orgao">
        <Select name="orgao" id="orgao">
          <option value="SEPLAG">SEPLAG</option>
          <option value="SEE">SEE</option>
        </Select>
      </FormField>

      <FormField label="Caixa do SEI" id="sei_unidade_id">
        <Select name="sei_unidade_id" id="sei_unidade_id">
          {#each data.unidades as unidade}
            <option value={unidade.id}>{unidade.sigla}</option>
          {/each}
        </Select>
      </FormField>
    </div>
  {/if}
{/if}
