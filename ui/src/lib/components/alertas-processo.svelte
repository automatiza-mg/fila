<script lang="ts">
  import type { ProcessoAposentadoria } from "$lib/api/types";
  import WarningIcon from "phosphor-svelte/lib/WarningIcon";

  type Props = {
    processo: ProcessoAposentadoria;
  };

  let { processo }: Props = $props();
</script>

{#if processo.alertas && processo.alertas.length > 0}
  <div
    role="alert"
    class="border border-red-600 bg-red-50 text-red-900 px-4 py-3 max-w-3xl rounded-sm text-sm"
  >
    <p class="flex items-center gap-1.5 font-medium">
      <WarningIcon class="size-4" />
      Alertas
    </p>
    <ul class="mt-1 list-disc pl-6 space-y-0.5">
      {#each processo.alertas as alerta}
        <li>{alerta}</li>
      {/each}
    </ul>
  </div>
{/if}
