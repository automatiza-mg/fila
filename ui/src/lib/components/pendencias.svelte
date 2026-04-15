<script lang="ts">
  import type { Papel, Pendencia } from "$lib/api/types";
  import { Popover } from "bits-ui";
  import AnalistaForm from "./analista-form.svelte";

  type Props = {
    usuarioId: number;
    pendencias: Pendencia[];
    papel?: Papel;
  };

  let { pendencias, usuarioId, papel }: Props = $props();
</script>

{#if pendencias.length === 0}
  <span>Não possui</span>
{:else}
  <Popover.Root>
    <Popover.Trigger class="underline">
      {pendencias.length} pendências
    </Popover.Trigger>
    <Popover.Content
      class="border border-border-strong shadow-sm z-30 max-w-80 rounded-xl p-4 w-full bg-surface space-y-1"
      sideOffset={8}
    >
      <p class="text-sm">Pendências do usuário:</p>
      <ul class="text-sm text-muted-foreground space-y-0.5 list-disc">
        {#each pendencias as pendencia}
          <li class="ml-6">
            <p class="tracking-tight">
              {pendencia.titulo}
              {#if pendencia.slug === "dados-analista"}
                - <AnalistaForm
                  {usuarioId}
                  buttonText="Resolver"
                  buttonVariant="link"
                />
              {/if}
            </p>
          </li>
        {/each}
      </ul>
      {#if papel === "ANALISTA"}
        <hr class="border-border mt-2" />
        <p class="text-xs text-muted-foreground mt-2">
          Todas as pendências precisam ser resolvidas antes do usuário começar
          a receber processos.
        </p>
      {/if}
    </Popover.Content>
  </Popover.Root>
{/if}
