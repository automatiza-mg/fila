<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";
  import { cn } from "$lib/cn";

  type Props = HTMLInputAttributes & { class?: string };

  let { class: className, ...rest }: Props = $props();
</script>

<input
  {...rest}
  class={cn(
    "rounded-lg border border-stone-200 bg-white p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50 read-only:bg-stone-50",
    className,
  )}
/>
