<script lang="ts">
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
      <div class="grid gap-1">
        <label for="orgao">Órgao de Exercício</label>
        <select name="orgao" id="orgao">
          <option value="SEPLAG">SEPLAG</option>
          <option value="SEE">SEE</option>
        </select>
      </div>

      <div class="grid gap-1">
        <label for="sei_unidade_id">Caixa do SEI</label>
        <select name="sei_unidade_id" id="sei_unidade_id">
          {#each data.unidades as unidade}
            <option value={unidade.id}>{unidade.sigla}</option>
          {/each}
        </select>
      </div>
    </div>
  {/if}
{/if}
