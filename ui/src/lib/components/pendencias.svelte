<script lang="ts">
  import type { Pendencia } from "$lib/api/types";
  import { Popover } from "bits-ui";

  type Props = {
    usuarioId: number;
    pendencias: Pendencia[];
  };

  let { pendencias, usuarioId }: Props = $props();
</script>

{#if pendencias.length === 0}
  <span>Não possui</span>
{:else}
  <Popover.Root>
    <Popover.Trigger class="decoration-dashed underline">
      {pendencias.length} pendências
    </Popover.Trigger>
    <Popover.Portal>
      <Popover.Content
        class="border border-stone-300 shadow-sm z-30 max-w-80 rounded-xl p-4 w-full bg-white space-y-1"
        sideOffset={8}
      >
        <p class="text-sm">Pendências do usuário:</p>
        <ul class="text-sm text-muted-foreground space-y-0.5 list-disc">
          {#each pendencias as pendencia}
            <li class="ml-6">
              <p class="tracking-tight">{pendencia.titulo}</p>
            </li>
          {/each}
        </ul>
      </Popover.Content>
    </Popover.Portal>
  </Popover.Root>
{/if}
