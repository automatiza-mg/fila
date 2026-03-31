<script lang="ts">
  import CopyIcon from "phosphor-svelte/lib/CopyIcon";
  import CheckIcon from "phosphor-svelte/lib/CheckIcon";
  import { Tooltip } from "bits-ui";

  type Props = {
    numero: string;
  };

  let { numero }: Props = $props();
  let copied = $state(false);

  let id: number;
  async function copy() {
    if (id) clearTimeout(id);
    await navigator.clipboard.writeText(numero);
    copied = true;
    id = setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="flex items-center gap-2 text-sm sm:text-base">
  <span>Nº processo SEI: <strong>{numero}</strong></span>
  <Tooltip.Provider>
    <Tooltip.Root>
      <Tooltip.Trigger
        onclick={copy}
        class="text-muted-foreground hover:text-foreground cursor-pointer p-1"
      >
        {#if copied}
          <CheckIcon class="size-4" />
        {:else}
          <CopyIcon class="size-4" />
        {/if}
      </Tooltip.Trigger>
      <Tooltip.Portal>
        <Tooltip.Content
          class="bg-stone-900/90 text-stone-50 text-xs rounded-md px-2 py-1 shadow-sm"
          sideOffset={4}
        >
          {copied ? "Copiado!" : "Copiar Número"}
        </Tooltip.Content>
      </Tooltip.Portal>
    </Tooltip.Root>
  </Tooltip.Provider>
</div>
