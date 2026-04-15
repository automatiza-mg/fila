<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
  import {
    getDiligenciaState,
    categoriasDiligencia,
  } from "$lib/stores/diligencia.svelte";
  import AlertDialog from "$lib/components/ui/alert-dialog.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Dialog from "$lib/components/ui/dialog.svelte";
  import Label from "$lib/components/ui/label.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import Textarea from "$lib/components/ui/textarea.svelte";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import ClipboardTextIcon from "phosphor-svelte/lib/ClipboardTextIcon";
  import PencilSimpleIcon from "phosphor-svelte/lib/PencilSimpleIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { Tooltip } from "bits-ui";

  const diligenciaStore = getDiligenciaState();
  let diligenciaForm = $state({
    tipo: "",
    subcategorias: [] as string[],
    detalhe: "",
  });

  let open = $state(false);
  let editingIndex = $state<number | null>(null);

  let categoriaAtual = $derived(
    categoriasDiligencia.find((c) => c.nome === diligenciaForm.tipo),
  );

  function resetDiligenciaForm() {
    diligenciaForm.tipo = "";
    diligenciaForm.subcategorias = [];
    diligenciaForm.detalhe = "";
    editingIndex = null;
  }

  function editDiligencia(index: number) {
    const d = diligenciaStore.diligencias[index];
    diligenciaForm.tipo = d.tipo;
    diligenciaForm.subcategorias = [...d.subcategorias];
    diligenciaForm.detalhe = d.detalhe;
    editingIndex = index;
    open = true;
  }

  function toggleSubcategoria(sub: string) {
    const idx = diligenciaForm.subcategorias.indexOf(sub);
    if (idx === -1) {
      diligenciaForm.subcategorias = [...diligenciaForm.subcategorias, sub];
    } else {
      diligenciaForm.subcategorias = diligenciaForm.subcategorias.filter(
        (s) => s !== sub,
      );
    }
  }
</script>

<svelte:head>
  <title>Solicitar Diligência - Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-8">
  <div class="flex items-center justify-between">
    <NumeroProcesso numero={data.processo.numero} />

    <a href="/analise" class="flex flex-col items-center">
      <ArrowElbowUpLeftIcon class="size-5" />
      <span>Voltar</span>
    </a>
  </div>

  <div class="mx-auto max-w-2xl space-y-6">
    <p
      class="text-center uppercase text-xl text-primary font-bold tracking-tight"
    >
      Diligências
    </p>

    <div>
      <Dialog
        buttonText="Adicionar Diligência"
        bind:open
        onOpenChange={(isOpen) => {
          if (!isOpen) resetDiligenciaForm();
        }}
      >
        {#snippet title()}
          {editingIndex !== null ? "Editar Diligência" : "Adicionar Diligência"}
        {/snippet}
        {#snippet description()}
          Preencha o tipo e informações da diligência
        {/snippet}

        <form
          class="flex flex-col gap-4 mt-6"
          onsubmit={(e) => {
            e.preventDefault();
            if (editingIndex !== null) {
              diligenciaStore.update(editingIndex, diligenciaForm);
            } else {
              diligenciaStore.add(diligenciaForm);
            }
            open = false;
            resetDiligenciaForm();
          }}
        >
          <Select
            required
            bind:value={diligenciaForm.tipo}
            onchange={() => {
              diligenciaForm.subcategorias = [];
              diligenciaForm.detalhe = "";
            }}
          >
            <option value="" selected>Selecione o tipo de diligência</option>
            {#each categoriasDiligencia as categoria}
              <option value={categoria.nome}>{categoria.nome}</option>
            {/each}
          </Select>

          {#if categoriaAtual?.subcategorias}
            <fieldset class="space-y-2">
              <legend class="text-sm font-medium">
                Selecione os documentos aplicáveis
              </legend>
              <div
                class="max-h-72 overflow-y-auto space-y-1.5 rounded-lg border border-border p-3"
              >
                {#each categoriaAtual.subcategorias as sub}
                  <Label class="flex items-start gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={diligenciaForm.subcategorias.includes(sub)}
                      onchange={() => toggleSubcategoria(sub)}
                      class="mt-0.5 size-4 rounded border-border-strong checked:border-primary checked:bg-primary text-primary focus:ring-secondary"
                    />
                    <span class="text-sm">{sub}</span>
                  </Label>
                {/each}
              </div>
            </fieldset>
          {/if}

          {#if !categoriaAtual?.subcategorias}
            <Textarea
              bind:value={diligenciaForm.detalhe}
              placeholder="Detalhar a diligência solicitada"
              rows={5}
            ></Textarea>
          {/if}

          <div class="flex justify-end">
            <Button>{editingIndex !== null ? "Salvar" : "Adicionar"}</Button>
          </div>
        </form>
      </Dialog>
    </div>

    {#if diligenciaStore.diligencias.length > 0}
      <ul class="space-y-2">
        {#each diligenciaStore.diligencias as diligencia, i}
          <li
            class="flex items-center gap-3 rounded-lg border border-border p-3"
          >
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium">
                {i + 1}. {diligencia.tipo}
              </p>
              {#if diligencia.subcategorias.length > 0}
                <p class="mt-0.5 text-xs text-muted-foreground">
                  {diligencia.subcategorias.length}
                  {diligencia.subcategorias.length === 1
                    ? "documento selecionado"
                    : "documentos selecionados"}
                </p>
              {/if}
              {#if diligencia.detalhe}
                <p class="mt-0.5 text-xs text-muted-foreground truncate">
                  {diligencia.detalhe}
                </p>
              {/if}
            </div>
            <Tooltip.Provider>
              <div class="flex items-center gap-1 shrink-0">
                <Tooltip.Root>
                  <Tooltip.Trigger
                    onclick={() => editDiligencia(i)}
                    class="text-primary bg-surface-alt hover:bg-surface-alt/80 cursor-pointer rounded-md p-1.5"
                  >
                    <PencilSimpleIcon class="size-4" />
                  </Tooltip.Trigger>
                  <Tooltip.Portal>
                    <Tooltip.Content
                      class="bg-tooltip-bg text-tooltip-fg text-xs rounded-md px-2 py-1 shadow-sm"
                      sideOffset={4}
                    >
                      Editar
                    </Tooltip.Content>
                  </Tooltip.Portal>
                </Tooltip.Root>

                <Tooltip.Root>
                  <Tooltip.Trigger
                    onclick={() => diligenciaStore.removeByIndex(i)}
                    class="text-destructive bg-surface-alt hover:bg-surface-alt/80 cursor-pointer rounded-md p-1.5"
                  >
                    <TrashIcon class="size-4" />
                  </Tooltip.Trigger>
                  <Tooltip.Portal>
                    <Tooltip.Content
                      class="bg-tooltip-bg text-tooltip-fg text-xs rounded-md px-2 py-1 shadow-sm"
                      sideOffset={4}
                    >
                      Excluir
                    </Tooltip.Content>
                  </Tooltip.Portal>
                </Tooltip.Root>
              </div>
            </Tooltip.Provider>
          </li>
        {/each}
      </ul>
    {:else}
      <div class="flex flex-col items-center gap-2 py-10 border border-border">
        <ClipboardTextIcon class="size-7" />
        <div>
          <p class="text-center font-semibold">
            Nenhuma diligência adicionada.
          </p>
          <p class="text-center text-sm text-muted-foreground">
            Utilize o botão acima para adicionar diligências.
          </p>
        </div>
      </div>
    {/if}

    <div class="flex justify-end">
      <AlertDialog
        buttonText="Enviar Diligência"
        disabled={diligenciaStore.diligencias.length === 0}
        onConfirmed={() => {}}
      >
        {#snippet title()}
          Confirmar Envio de Diligência
        {/snippet}
        {#snippet description()}
          Revise as diligências antes de enviar.
        {/snippet}

        <ul class="mt-4 max-h-72 overflow-y-auto space-y-3 text-sm">
          {#each diligenciaStore.diligencias as diligencia, i}
            <li>
              <p class="font-medium">{i + 1}. {diligencia.tipo}</p>
              {#if diligencia.subcategorias.length > 0}
                <ul
                  class="ml-4 mt-1 list-disc text-muted-foreground space-y-0.5"
                >
                  {#each diligencia.subcategorias as sub}
                    <li>{sub}</li>
                  {/each}
                </ul>
              {/if}
              {#if diligencia.detalhe}
                <p class="ml-4 mt-1 text-muted-foreground">
                  {diligencia.detalhe}
                </p>
              {/if}
            </li>
          {/each}
        </ul>
      </AlertDialog>
    </div>
  </div>
</div>
