<script lang="ts">
  import AnalistaForm from "$lib/components/analista-form.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import { toast } from "svelte-sonner";
  import { deleteUsuarioCmd, enviarCadastroCmd } from "../usuario.remote";
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

<!-- Reenviar Email de Cadastro -->
{#if !data.usuario.email_verificado}
  <Button
    onclick={async () => {
      try {
        await enviarCadastroCmd({ usuarioId: data.usuario.id });
        toast.success("Email de cadastro reenviado com sucesso!");
      } catch {
        toast.error("Não foi possível reenviar o email de cadastro");
      }
    }}
  >
    Reenviar Email de Cadastro
  </Button>
{/if}

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
    <AnalistaForm usuarioId={data.usuario.id} />
  {/if}
{/if}
