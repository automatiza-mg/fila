<script lang="ts">
  import type { HTMLLabelAttributes } from "svelte/elements";
  import type { Snippet } from "svelte";

  type Props = HTMLLabelAttributes & {
    children: Snippet;
  };

  let { children, ...rest }: Props = $props();
</script>

<label class="font-medium w-fit" {...rest}>
  {@render children()}
</label>
