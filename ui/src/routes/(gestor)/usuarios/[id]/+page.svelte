<script lang="ts">
  import FormField from "$lib/components/ui/form-field.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import { toast } from "svelte-sonner";
  import { deleteUsuarioCmd } from "../usuario.remote";
  import { goto } from "$app/navigation";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import type { PageProps } from "./$types";
  import AlertDialog from "$lib/components/ui/alert-dialog.svelte";

  let { data }: PageProps = $props();
  let open = $state(false);
</script>

<svelte:head>
  <title>{data.usuario.nome} - Fila Aposentadoria</title>
</svelte:head>

<div class="flex justify-end">
  <a href="/usuarios" class="flex flex-col items-center">
    <ArrowElbowUpLeftIcon class="size-5" />
    <span>Voltar</span>
  </a>
</div>

<div>
  Nome: {data.usuario.nome}
</div>

<!-- Excluir -->
{#if data.usuario.id !== data.usuarioAtual.id}
  <AlertDialog
    bind:open
    buttonText="Exlcuir Usuário"
    variant="destructive"
    onConfirmed={async () => {
      try {
        await deleteUsuarioCmd({ usuarioId: data.usuario.id });
        toast.success("Usuário excluído com sucesso!");
        goto("/usuarios");
      } catch (err) {
        toast.error("Não foi possível excluir o usuário");
      }
    }}
  >
    {#snippet title()}
      Excluir Usuário
    {/snippet}
    {#snippet description()}
      Essa é uma ação irreversível e não poderá ser desfeita. Tem certeza que
      deseja continuar?
    {/snippet}
  </AlertDialog>
{/if}

{#if data.usuario.papel === "ANALISTA"}
  {#if data.analista}
    Dados Analista
  {:else}
    <div class="flex flex-col gap-4 max-w-sm">
      <FormField label="Órgao de Exercício" id="orgao">
        <Select name="orgao" id="orgao">
          <option value="SEPLAG">SEPLAG</option>
          <option value="SEE">SEE</option>
        </Select>
      </FormField>

      <FormField label="Caixa do SEI" id="sei_unidade_id">
        <Select name="sei_unidade_id" id="sei_unidade_id">
          {#each data.unidades as unidade}
            <option value={unidade.id}>{unidade.sigla}</option>
          {/each}
        </Select>
      </FormField>
    </div>
  {/if}
{/if}
