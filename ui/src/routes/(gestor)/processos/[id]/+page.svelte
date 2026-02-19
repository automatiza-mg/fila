<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<a href="/processos">Ver Processos</a>

<pre class="text-xs">{JSON.stringify(data.processo, null, 4)}</pre>
