<script lang="ts">
  import CircleCheck from "@lucide/svelte/icons/circle-check";
  import Circle from "@lucide/svelte/icons/circle";

  let { password = "" }: { password: string } = $props();

  let rules = $derived([
    { label: "Pelo menos 8 caracteres", fulfilled: password.length >= 8 },
    {
      label: "No maximo 60 caracteres",
      fulfilled: password.length > 0 && password.length <= 60,
    },
    { label: "Pelo menos um digito", fulfilled: /\d/.test(password) },
    {
      label: "Pelo menos um caractere especial",
      fulfilled: /[^\w\s]/.test(password),
    },
  ]);
</script>

<ul class="flex flex-col gap-1">
  {#each rules as rule}
    <li
      class="flex items-center gap-2 text-sm {rule.fulfilled
        ? 'text-green-600'
        : 'text-stone-400'}"
    >
      {#if rule.fulfilled}
        <CircleCheck size={16} />
      {:else}
        <Circle size={16} />
      {/if}
      {rule.label}
    </li>
  {/each}
</ul>
