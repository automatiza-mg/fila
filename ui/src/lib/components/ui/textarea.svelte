<script lang="ts">
  import type { HTMLTextareaAttributes } from "svelte/elements";

  type Props = HTMLTextareaAttributes;

  let { ...rest }: Props = $props();
</script>

<textarea
  {...rest}
  class="p-2 rounded-xl border border-stone-200 focus-visible:ring-3 outline-none focus-visible:ring-secondary/50 focus-visible:border-secondary read-only:bg-stone-50"
></textarea>
