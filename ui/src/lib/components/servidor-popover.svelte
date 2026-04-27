<script lang="ts">
  import { Popover } from "bits-ui";
  import { formatCpf } from "$lib/formatter";
  import ServidorPopoverContent from "./servidor-popover-content.svelte";

  type Props = {
    cpf: string;
  };

  let { cpf }: Props = $props();

  let open = $state(false);
</script>

<Popover.Root bind:open>
  <Popover.Trigger class="underline cursor-pointer">
    {formatCpf(cpf)}
  </Popover.Trigger>
  <Popover.Content
    class="border border-border-strong shadow-md z-30 w-72 rounded-xl p-4 bg-surface"
    sideOffset={8}
  >
    {#if open}
      <ServidorPopoverContent {cpf} />
    {/if}
  </Popover.Content>
</Popover.Root>
