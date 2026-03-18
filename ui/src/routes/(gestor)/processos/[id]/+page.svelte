<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processo {data.processo.numero} - Fila Aposentadoria</title>
</svelte:head>

<div>
  <span>Nº processo SEI:</span>
  <span>{data.processo.numero}</span>
</div>
