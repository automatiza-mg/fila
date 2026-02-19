<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Cadastrar | Fila Aposentadoria</title>
</svelte:head>

<pre class="text-xs">{JSON.stringify(data.usuario, null, 4)}</pre>
