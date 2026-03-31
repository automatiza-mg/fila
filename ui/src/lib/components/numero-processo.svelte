<script lang="ts">
  import CopyIcon from "phosphor-svelte/lib/CopyIcon";
  import CheckIcon from "phosphor-svelte/lib/CheckIcon";

  type Props = {
    numero: string;
  };

  let { numero }: Props = $props();
  let copied = $state(false);

  async function copy() {
    await navigator.clipboard.writeText(numero);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="flex items-center gap-2 text-sm">
  <span>Nº processo SEI: <strong>{numero}</strong></span>
  <button
    onclick={copy}
    class="text-muted-foreground hover:text-foreground"
    title="Copiar número"
  >
    {#if copied}
      <CheckIcon class="size-4" />
    {:else}
      <CopyIcon class="size-4" />
    {/if}
  </button>
</div>
