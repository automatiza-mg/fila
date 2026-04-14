<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import ProcessoInfo from "$lib/components/processo-info.svelte";
  import LeituraInvalidaDialog from "$lib/components/leitura-invalida-dialog.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import ArrowSquareOutIcon from "phosphor-svelte/lib/ArrowSquareOutIcon";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Analista - Fila Aposentadoria</title>
</svelte:head>

{#if data.processo}
  <div class="space-y-8">
    <div
      class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
    >
      <NumeroProcesso numero={data.processo.numero} />
      {#if data.processo.possui_preview}
        <Button
          variant="outline"
          href="/preview/{data.processo.id}"
          target="_blank"
          rel="noopener noreferrer"
        >
          Pré-visualizar Processo
          <ArrowSquareOutIcon />
        </Button>
      {/if}
    </div>

    <ProcessoInfo processo={data.processo} />

    <div class="flex gap-2">
      <Button href="/analise/diligencia">Solicitar Diligência</Button>
      <LeituraInvalidaDialog processoId={data.processo.id} />
    </div>
  </div>
{:else}
  <div class="flex grow items-center justify-center">
    <p class="text-muted-foreground text-sm">
      Nenhum processo atribuído no momento.
    </p>
  </div>
{/if}
