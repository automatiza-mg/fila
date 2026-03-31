<script lang="ts">
  import type { HTMLTextareaAttributes } from "svelte/elements";
  import { cn } from "$lib/cn";

  type Props = HTMLTextareaAttributes & { class?: string };

  let { class: className, ...rest }: Props = $props();
</script>

<textarea
  {...rest}
  class={cn(
    "rounded-lg border border-stone-200 bg-white p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50 read-only:bg-stone-50",
    className,
  )}
></textarea>
