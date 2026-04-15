<script lang="ts">
  import { LinkPreview } from "bits-ui";
  import { formatCpf } from "$lib/formatter";
  import SpinnerGapIcon from "phosphor-svelte/lib/SpinnerGapIcon";
  import { servidorQuery } from "../../routes/(gestor)/processos/processo.remote";

  type Props = {
    cpf: string;
  };

  let { cpf }: Props = $props();

  // svelte-ignore state_referenced_locally
  const servidor = servidorQuery({ cpf });

</script>

<LinkPreview.Root openDelay={400} closeDelay={200}>
  <LinkPreview.Trigger class="underline cursor-pointer">
    {#snippet child({ props })}
      <span {...props}>{formatCpf(cpf)}</span>
    {/snippet}
  </LinkPreview.Trigger>
  <LinkPreview.Portal>
    <LinkPreview.Content
      class="border border-border-strong shadow-md z-30 w-72 rounded-xl p-4 bg-surface"
      sideOffset={8}
    >
      {#if !servidor.ready}
        <div class="h-32 flex items-center justify-center">
          <p class="text-sm text-muted-foreground flex items-center gap-1.5">
            <SpinnerGapIcon class="size-4 animate-spin" />
            Carregando...
          </p>
        </div>
      {:else if servidor.current === null}
        <p class="text-sm text-muted-foreground">
          Servidor não encontrado para este CPF.
        </p>
      {:else}
        {@const s = servidor.current}
        <p class="text-sm font-semibold">{s.nome}</p>
        <p class="text-xs text-muted-foreground">
          {s.sexo === "M" ? "Masculino" : "Feminino"}, nascido(a) em {new Date(
            s.data_nascimento,
          ).toLocaleDateString("pt-BR", { timeZone: "UTC" })}
        </p>
        <div class="text-sm text-muted-foreground space-y-0.5 mt-2">
          <p>MASP: {s.masp}</p>
          <p>CPF: {formatCpf(s.cpf)}</p>
          {#if s.possui_deficiencia}
            <p>Possui deficiência: Sim</p>
          {/if}
        </div>
        <hr class="border-border mt-3" />
        <p class="text-xs text-muted-foreground mt-2">
          Dados consultados no Data Lake do SISAP
        </p>
      {/if}
    </LinkPreview.Content>
  </LinkPreview.Portal>
</LinkPreview.Root>
