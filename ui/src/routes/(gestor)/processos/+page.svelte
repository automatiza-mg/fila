<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processos | Fila Aposentadoria</title>
</svelte:head>

<ul>
  {#each data.processos?.data as processo}
    <li>
      <a href="/processos/{processo.id}">{processo.numero}</a> | {processo.status}
    </li>
  {/each}
</ul>
