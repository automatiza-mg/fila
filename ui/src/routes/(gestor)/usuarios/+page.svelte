<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Usuários | Fila Aposentadoria</title>
</svelte:head>

<div class="flex justify-end">
  <button
    class="px-4 py-2 font-semibold bg-primary text-white rounded-xl border border-transparent"
  >
    Novo Cadastro
  </button>
</div>
