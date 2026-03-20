<script lang="ts">
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

  function prioridadeStatus(status: string) {
    switch (status) {
      case "pendente":
        return "Pendente";
      case "negado":
        return "Negado";
      case "aprovado":
        return "Aprovado";
      default:
        return "Desconhecido";
    }
  }
</script>

<svelte:head>
  <title>Solicitações de Prioridade - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex justify-end">
    <a href="/processos" class="flex flex-col items-center">
      <ArrowElbowUpLeftIcon class="size-5" />
      <span>Voltar</span>
    </a>
  </div>

  <div>
    <table class="w-full border border-stone-200 text-sm">
      <thead>
        <tr class="border-y border-stone-200 bg-stone-100">
          <th scope="col" class="text-left font-semibold p-2.5">
            Número Processo
          </th>
          <th scope="col" class="text-left font-semibold p-2.5">
            Justificativa
          </th>
          <th scope="col" class="text-left font-semibold p-2.5">
            Data Solicitação
          </th>
          <th scope="col" class="text-left font-semibold p-2.5"> Status </th>
        </tr>
      </thead>
      <tbody>
        {#each data.solicitacoes.data as solicitacao}
          <tr class="border-b border-stone-200 hover:bg-stone-50">
            <td class="p-2.5">
              <a
                href={`/processos/${solicitacao.processo_aposentadoria_id}`}
                class="underline text-primary"
              >
                {solicitacao.numero_processo}
              </a>
            </td>
            <td class="p-2.5">
              {solicitacao.justificativa}
            </td>
            <td class="p-2.5">
              {new Date(solicitacao.criado_em).toLocaleDateString()}
            </td>
            <td class="p-2.5">
              {prioridadeStatus(solicitacao.status)}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
