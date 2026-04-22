<script lang="ts">
  import AnalistaForm from "$lib/components/analista-form.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import AlertDialog from "$lib/components/ui/alert-dialog.svelte";
  import { formatCpf } from "$lib/formatter";
  import { toast } from "svelte-sonner";
  import {
    afastarAnalistaCmd,
    deleteUsuarioCmd,
    enviarCadastroCmd,
    retornarAnalistaCmd,
  } from "$lib/fns/usuarios.remote";
  import { goto, invalidateAll } from "$app/navigation";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import EnvelopeIcon from "phosphor-svelte/lib/EnvelopeIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import UserMinusIcon from "phosphor-svelte/lib/UserMinusIcon";
  import UserPlusIcon from "phosphor-svelte/lib/UserPlusIcon";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
  let deleteOpen = $state(false);
  let afastarOpen = $state(false);
  let retornarOpen = $state(false);
</script>

<svelte:head>
  <title>{data.usuario.nome} - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex justify-end items-center">
    <a href="/usuarios" class="flex flex-col items-center">
      <ArrowElbowUpLeftIcon class="size-5" />
      <span>Voltar</span>
    </a>
  </div>

  <div class="space-y-3">
    <div>
      <h2 class="text-xl font-bold">Informações Gerais</h2>
      <p class="text-sm text-muted-foreground">
        Dados cadastrais do usuário no sistema.
      </p>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 max-w-4xl">
      <FormField label="Nome" id="nome">
        <Input type="text" readonly id="nome" value={data.usuario.nome} />
      </FormField>

      <FormField label="CPF" id="cpf">
        <Input
          type="text"
          readonly
          id="cpf"
          value={formatCpf(data.usuario.cpf)}
        />
      </FormField>

      <FormField label="Email" id="email">
        <Input type="text" readonly id="email" value={data.usuario.email} />
      </FormField>

      <FormField label="Papel" id="papel">
        <Input
          type="text"
          readonly
          id="papel"
          value={data.usuario.papel ?? "Não definido"}
        />
      </FormField>

      <FormField label="Email Verificado" id="email-verificado">
        <Input
          type="text"
          readonly
          id="email-verificado"
          value={data.usuario.email_verificado ? "Sim" : "Não"}
        />
      </FormField>
    </div>
  </div>

  <!-- Dados Analista -->
  {#if data.usuario.papel === "ANALISTA"}
    <div class="space-y-3">
      <div>
        <h2 class="text-base font-semibold">Dados do Analista</h2>
        <p class="text-sm text-muted-foreground">
          Informações complementares para a atuação como analista de processos.
        </p>
      </div>

      {#if data.analista}
        <div
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 max-w-4xl"
        >
          <FormField label="Órgão de Exercício" id="orgao">
            <Input
              type="text"
              readonly
              id="orgao"
              value={data.analista.orgao}
            />
          </FormField>

          <FormField label="Unidade SEI" id="unidade-sei">
            <Input
              type="text"
              readonly
              id="unidade-sei"
              value={data.analista.sei_unidade_sigla}
            />
          </FormField>

          <FormField label="Afastado" id="afastado">
            <Input
              type="text"
              readonly
              id="afastado"
              value={data.analista.afastado ? "Sim" : "Não"}
            />
          </FormField>
        </div>
      {:else}
        <p class="text-sm text-muted-foreground">
          Dados de analista ainda não cadastrados.
        </p>
        <AnalistaForm usuarioId={data.usuario.id} />
      {/if}
    </div>
  {/if}

  <!-- Ações do Usuário -->
  <div class="space-y-3">
    <div>
      <h2 class="text-base font-semibold">Ações do Usuário</h2>
      <p class="text-sm text-muted-foreground">
        Gerenciamento de conta e permissões do usuário.
      </p>
    </div>

    <div class="flex flex-wrap items-center gap-4">
      {#if !data.usuario.email_verificado}
        <Button
          variant="outline"
          onclick={async () => {
            try {
              await enviarCadastroCmd({ usuarioId: data.usuario.id });
              toast.success("Email de cadastro reenviado com sucesso!");
            } catch {
              toast.error("Não foi possível reenviar o email de cadastro");
            }
          }}
        >
          <EnvelopeIcon />
          Reenviar Email de Cadastro
        </Button>
      {/if}

      {#if data.analista}
        {#if data.analista.afastado}
          <AlertDialog
            bind:open={retornarOpen}
            buttonText="Retornar Analista"
            onConfirmed={async () => {
              try {
                await retornarAnalistaCmd({ usuarioId: data.usuario.id });
                retornarOpen = false;
                toast.success("Analista retornado com sucesso!");
                await invalidateAll();
              } catch {
                toast.error("Não foi possível retornar o analista");
              }
            }}
          >
            {#snippet buttonIcon()}<UserPlusIcon />{/snippet}
            {#snippet title()}
              Retornar Analista
            {/snippet}
            {#snippet description()}
              O analista voltará a receber processos para análise. Deseja
              continuar?
            {/snippet}
          </AlertDialog>
        {:else}
          <AlertDialog
            bind:open={afastarOpen}
            buttonText="Afastar Analista"
            variant="outline"
            onConfirmed={async () => {
              try {
                await afastarAnalistaCmd({ usuarioId: data.usuario.id });
                afastarOpen = false;
                toast.success("Analista afastado com sucesso!");
                await invalidateAll();
              } catch {
                toast.error("Não foi possível afastar o analista");
              }
            }}
          >
            {#snippet buttonIcon()}<UserMinusIcon />{/snippet}
            {#snippet title()}
              Afastar Analista
            {/snippet}
            {#snippet description()}
              O analista deixará de receber novos processos enquanto estiver
              afastado. Deseja continuar?
            {/snippet}
          </AlertDialog>
        {/if}
      {/if}

      {#if data.usuario.id !== data.usuarioAtual.id}
        <AlertDialog
          bind:open={deleteOpen}
          buttonText="Excluir Usuário"
          variant="destructive"
          onConfirmed={async () => {
            try {
              await deleteUsuarioCmd({ usuarioId: data.usuario.id });
              toast.success("Usuário excluído com sucesso!");
              goto("/usuarios");
            } catch {
              toast.error("Não foi possível excluir o usuário");
            }
          }}
        >
          {#snippet buttonIcon()}<TrashIcon />{/snippet}
          {#snippet title()}
            Excluir Usuário
          {/snippet}
          {#snippet description()}
            Essa é uma ação irreversível e não poderá ser desfeita. Tem certeza
            que deseja continuar?
          {/snippet}
        </AlertDialog>
      {/if}
    </div>
  </div>
</div>
