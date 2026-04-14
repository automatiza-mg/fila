<script lang="ts">
  import type { HTMLTextareaAttributes } from "svelte/elements";
  import { cn } from "$lib/cn";

  type Props = HTMLTextareaAttributes & { class?: string };

  let { class: className, value = $bindable(), ...rest }: Props = $props();
</script>

<textarea
  bind:value
  {...rest}
  class={cn(
    "rounded-lg border border-border bg-surface p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50 read-only:bg-surface-subtle",
    className,
  )}
></textarea>
