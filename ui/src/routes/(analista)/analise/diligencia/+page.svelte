<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
  import {
    getDiligenciaState,
    categoriasDiligencia,
  } from "$lib/stores/diligencia.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Dialog from "$lib/components/ui/dialog.svelte";
  import Label from "$lib/components/ui/label.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import Textarea from "$lib/components/ui/textarea.svelte";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
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
        <li class="text-sm flex items-center gap-2">
          <div>
            <span>{i + 1}- {diligencia.tipo}</span>
            {#if diligencia.subcategorias.length > 0}
              <span class="text-muted-foreground">
                ({diligencia.subcategorias.length}
                {diligencia.subcategorias.length === 1
                  ? "documento"
                  : "documentos"})
              </span>
            {/if}
          </div>
          <Tooltip.Provider>
            <Tooltip.Root>
              <Tooltip.Trigger
                onclick={() => editDiligencia(i)}
                class="text-primary bg-stone-100 hover:bg-stone-200 cursor-pointer rounded-md p-1.5"
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
                class="text-destructive bg-stone-100 hover:bg-stone-200 cursor-pointer rounded-md p-1.5"
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
          </Tooltip.Provider>
        </li>
      {/each}
    </ul>
  {/if}
</div>
