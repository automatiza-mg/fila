<script lang="ts">
  import { statusProcesso } from "$lib/processo";
  import { process } from "zod/v4/core";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Processos - Fila Aposentadoria</title>
</svelte:head>

<div>
  <table class="w-full border border-stone-200 text-sm">
    <thead>
      <tr class="border-y border-stone-200 bg-stone-100">
        <th scope="col" class="text-left font-semibold p-2.5">Numero</th>
        <th scope="col" class="text-left font-semibold p-2.5">Status</th>
        <th scope="col" class="text-left font-semibold p-2.5">
          Data Requerimento
        </th>
        <th scope="col" class="text-left font-semibold p-2.5">Score</th>
        <th scope="col" class="text-left font-semibold p-2.5">Analista</th>
      </tr>
    </thead>
    <tbody>
      {#each data.processos.data as processo}
        <tr class="border-b border-stone-200 hover:bg-stone-50">
          <td class="p-2.5">
            <a
              class="text-primary underline"
              href={`/processos/${processo.id}`}
            >
              {processo.numero}</a
            >
          </td>
          <td class="p-2.5">{statusProcesso(processo.status)}</td>
          <td class="p-2.5"
            >{new Date(processo.data_requerimento).toLocaleDateString("pt-BR", {
              timeZone: "UTC",
            })}</td
          >
          <td class="p-2.5">{processo.score}</td>
          <td class="p-2.5">{processo.analista ?? "Não possui"}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
