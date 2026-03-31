<script lang="ts">
  import type { HTMLSelectAttributes } from "svelte/elements";
  import type { Snippet } from "svelte";
  import { cn } from "$lib/cn";

  type Props = HTMLSelectAttributes & {
    children: Snippet;
    class?: string;
  };

  let { children, class: className, ...rest }: Props = $props();
</script>

<select
  {...rest}
  class={cn(
    "rounded-lg border border-stone-200 bg-white p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50",
    className,
  )}
>
  {@render children()}
</select>
