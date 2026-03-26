<script lang="ts">
  import type { HTMLSelectAttributes } from "svelte/elements";
  import type { Snippet } from "svelte";

  type Props = HTMLSelectAttributes & {
    children: Snippet;
  };

  let { children, ...rest }: Props = $props();
</script>

<select
  {...rest}
  class="p-2 text-sm rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary"
>
  {@render children()}
</select>
