<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  let { value = $bindable(), ...props }: HTMLInputAttributes = $props();
</script>

<input
  class="py-2 px-3 border rounded-2xl border-stone-200 w-full read-only:bg-stone-100"
  bind:value
  {...props}
/>
