<script lang="ts">
  import type { HTMLSelectAttributes } from "svelte/elements";
  import type { Snippet } from "svelte";
  import { cn } from "$lib/cn";

  type Props = HTMLSelectAttributes & {
    children: Snippet;
    class?: string;
  };

  let {
    children,
    class: className,
    value = $bindable(),
    ...rest
  }: Props = $props();
</script>

<select
  bind:value
  {...rest}
  class={cn(
    "rounded-lg border border-border bg-surface p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50",
    className,
  )}
>
  {@render children()}
</select>
