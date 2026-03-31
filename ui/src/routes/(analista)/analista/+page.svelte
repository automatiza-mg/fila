<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Analista - Fila Aposentadoria</title>
</svelte:head>

{#if data.processo}
  <NumeroProcesso numero={data.processo.numero} />
{:else}
  <div class="flex grow items-center justify-center">
    <p class="text-muted-foreground text-sm">
      Nenhum processo atribuído no momento.
    </p>
  </div>
{/if}
