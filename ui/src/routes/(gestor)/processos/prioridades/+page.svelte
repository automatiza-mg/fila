<script lang="ts">
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import { hasPapel } from "$lib/papel";
  import { toast } from "svelte-sonner";
  import { invalidateAll } from "$app/navigation";
  import {
    aprovarPrioridadeCmd,
    negarPrioridadeCmd,
  } from "./prioridade.remote";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

  let isSubsecretario = $derived(hasPapel(data.usuario, "SUBSECRETARIO"));

  async function handleStatusChange(
    event: Event & { currentTarget: HTMLSelectElement },
    solicitacaoId: number,
  ) {
    const select = event.currentTarget;
    const valor = select.value;

    try {
      if (valor === "aprovado") {
        await aprovarPrioridadeCmd({ id: solicitacaoId });
        toast.success("Solicitação aprovada com sucesso");
      } else if (valor === "negado") {
        await negarPrioridadeCmd({ id: solicitacaoId });
        toast.success("Solicitação negada com sucesso");
      }
      await invalidateAll();
    } catch {
      select.value = "pendente";
      toast.error("Não foi possível atualizar a solicitação");
    }
  }
</script>

<svelte:head>
  <title>Solicitações de Prioridade - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <form method="GET" class="flex items-center gap-2">
      <label for="numero-filter" class="text-sm font-medium">Processo:</label>
      <input
        id="numero-filter"
        name="numero"
        type="text"
        value={data.numero}
        placeholder="Número do processo"
        class="rounded border border-stone-200 bg-white p-2 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50"
      />
      <label for="status-filter" class="text-sm font-medium">Status:</label>
      <select
        id="status-filter"
        name="status"
        class="rounded border border-stone-200 bg-white p-2 min-w-30 text-sm outline-none focus-visible:border-secondary focus-visible:ring-3 focus-visible:ring-secondary/50"
      >
        <option value="" selected={data.status === ""}>Todos</option>
        <option value="pendente" selected={data.status === "pendente"}
          >Pendente</option
        >
        <option value="aprovado" selected={data.status === "aprovado"}
          >Aprovado</option
        >
        <option value="negado" selected={data.status === "negado"}
          >Negado</option
        >
      </select>
      <button
        type="submit"
        class="rounded bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary/90"
      >
        Buscar
      </button>
    </form>
    <a href="/processos" class="flex flex-col items-center">
      <ArrowElbowUpLeftIcon class="size-5" />
      <span>Voltar</span>
    </a>
  </div>

  <div>
    {#if data.solicitacoes.data.length === 0}
      <div class="p-6 border border-stone-300">
        <p class="text-center font-medium">
          Nenhuma solicitação de prioridade encontrada.
        </p>
        <p class="text-center text-sm text-muted-foreground">
          Acesse os detalhes de um processo para criar uma nova prioridade.
        </p>
      </div>
    {:else}
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
                <select
                  disabled={!isSubsecretario}
                  class="rounded bg-white p-2 text-sm border border-stone-200 focus-visible:ring-3 outline-none disabled:bg-stone-100 focus-visible:ring-secondary/50 focus-visible:border-secondary w-full"
                  onchange={(e) => handleStatusChange(e, solicitacao.id)}
                >
                  <option
                    value="pendente"
                    selected={solicitacao.status === "pendente"}
                    disabled>Pendente</option
                  >
                  <option
                    value="aprovado"
                    selected={solicitacao.status === "aprovado"}
                    >Aprovado</option
                  >
                  <option
                    value="negado"
                    selected={solicitacao.status === "negado"}>Negado</option
                  >
                </select>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>
