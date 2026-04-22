<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import ProcessoInfo from "$lib/components/processo-info.svelte";
  import LeituraInvalidaDialog from "$lib/components/leitura-invalida-dialog.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import ArrowSquareOutIcon from "phosphor-svelte/lib/ArrowSquareOutIcon";
  import { onDestroy } from "svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

  let pollId: ReturnType<typeof setInterval> | null = null;

  $effect(() => {
    if (!data.processo) {
      pollId = setInterval(() => {
        invalidateAll();
      }, 2000);
    }

    return () => {
      if (pollId) {
        clearInterval(pollId);
        pollId = null;
      }
    };
  });

  onDestroy(() => {
    if (pollId) clearInterval(pollId);
  });
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
      <Button href="/analise/diligencia">Registrar Diligência</Button>
      <LeituraInvalidaDialog processoId={data.processo.id} />
    </div>
  </div>
{:else}
  <div class="flex flex-col grow items-center justify-center">
    <div class="flex max-w-md flex-col items-center text-center">
      <p class="font-semibold">Nenhum processo atribuído no momento.</p>
      <p class="text-muted-foreground text-xs">
        Assim que um processo estiver disponível, ele será atribuído a você
        automaticamente.
      </p>
    </div>
  </div>
{/if}
