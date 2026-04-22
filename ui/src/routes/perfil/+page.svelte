<script lang="ts">
  import AlterarSenhaDialog from "$lib/components/alterar-senha-dialog.svelte";
  import SairButton from "$lib/components/sair-button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { formatCpf } from "$lib/formatter";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();

  const papelLabels: Record<string, string> = {
    ADMIN: "Administrador",
    GESTOR: "Gestor",
    SUBSECRETARIO: "Subsecretário",
    ANALISTA: "Analista",
  };

  const papelLabel = $derived(
    data.usuario.papel
      ? (papelLabels[data.usuario.papel] ?? data.usuario.papel)
      : "Não definido",
  );
</script>

<svelte:head>
  <title>Meu perfil - Fila Aposentadoria</title>
</svelte:head>

<div class="from-secondary to-primary flex min-h-svh flex-col bg-linear-to-b">
  <header class="bg-surface flex p-6 justify-between items-center">
    <p class="text-xl font-bold">Fila Aposentadoria</p>
    <SairButton />
  </header>

  <main class="flex grow flex-col p-1.5 sm:p-4">
    <section class="mx-auto flex w-full max-w-4xl grow flex-col p-1.5 sm:p-4">
      <div
        class="bg-surface flex h-full grow flex-col rounded-2xl p-6 sm:p-8 shadow border border-secondary space-y-6"
      >
        <div class="flex items-center justify-between">
          <h1 class="text-2xl font-bold">Meu perfil</h1>
          <a href="/" class="flex flex-col items-center">
            <ArrowElbowUpLeftIcon class="size-5" />
            <span class="text-sm">Voltar</span>
          </a>
        </div>

        <div class="space-y-3">
          <div>
            <h2 class="text-lg font-semibold">Informações pessoais</h2>
            <p class="text-sm text-muted-foreground">
              Seus dados cadastrais no sistema.
            </p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
              <Input
                type="text"
                readonly
                id="email"
                value={data.usuario.email}
              />
            </FormField>

            <FormField label="Papel" id="papel">
              <Input type="text" readonly id="papel" value={papelLabel} />
            </FormField>
          </div>
        </div>

        {#if data.analista}
          <div class="space-y-3 border-t border-border pt-6">
            <div>
              <h2 class="text-lg font-semibold">Informações de analista</h2>
              <p class="text-sm text-muted-foreground">
                Dados complementares vinculados ao seu papel de analista.
              </p>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <FormField label="Órgão" id="orgao">
                <Input
                  type="text"
                  readonly
                  id="orgao"
                  value={data.analista.orgao}
                />
              </FormField>

              <FormField label="Unidade SEI" id="sei_unidade">
                <Input
                  type="text"
                  readonly
                  id="sei_unidade"
                  value={data.analista.sei_unidade_sigla}
                />
              </FormField>

              <FormField label="Situação" id="situacao">
                <Input
                  type="text"
                  readonly
                  id="situacao"
                  value={data.analista.afastado ? "Afastado" : "Ativo"}
                />
              </FormField>

              <FormField label="Última atribuição" id="ultima_atribuicao">
                <Input
                  type="text"
                  readonly
                  id="ultima_atribuicao"
                  value={data.analista.ultima_atribuicao_em
                    ? new Date(
                        data.analista.ultima_atribuicao_em,
                      ).toLocaleString("pt-BR")
                    : "Sem atribuições"}
                />
              </FormField>
            </div>
          </div>
        {/if}

        <div class="space-y-3 border-t border-border pt-6">
          <div>
            <h2 class="text-lg font-semibold">Segurança</h2>
            <p class="text-sm text-muted-foreground">
              Altere sua senha de acesso quando necessário.
            </p>
          </div>

          <AlterarSenhaDialog />
        </div>
      </div>
    </section>
  </main>
</div>
