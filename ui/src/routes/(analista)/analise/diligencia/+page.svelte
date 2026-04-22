<script lang="ts">
  import { goto, invalidateAll } from "$app/navigation";
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import AlertDialog from "$lib/components/ui/alert-dialog.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Dialog from "$lib/components/ui/dialog.svelte";
  import Label from "$lib/components/ui/label.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import Textarea from "$lib/components/ui/textarea.svelte";
  import { categoriasDiligencia } from "$lib/stores/diligencia.svelte";
  import { Popover, Tooltip } from "bits-ui";
  import ArrowElbowUpLeftIcon from "phosphor-svelte/lib/ArrowElbowUpLeftIcon";
  import PencilSimpleIcon from "phosphor-svelte/lib/PencilSimpleIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { toast } from "svelte-sonner";
  import type { PageProps } from "./$types";
  import {
    descartarDiligencia,
    enviarDiligencia,
    salvarDiligencia,
  } from "$lib/fns/diligencias.remote";

  let { data }: PageProps = $props();

  type ItemPayload = {
    tipo: string;
    subcategorias: string[];
    detalhe: string;
  };

  let diligenciaForm = $state({
    tipo: "",
    subcategorias: [] as string[],
    detalhe: "",
  });

  let open = $state(false);
  let editingId = $state<number | null>(null);
  let isSubmitting = $state(false);

  let itens = $derived(data.rascunho.itens);

  let categoriaAtual = $derived(
    categoriasDiligencia.find((c) => c.nome === diligenciaForm.tipo),
  );

  function resetDiligenciaForm() {
    diligenciaForm.tipo = "";
    diligenciaForm.subcategorias = [];
    diligenciaForm.detalhe = "";
    editingId = null;
  }

  function editDiligencia(id: number) {
    const d = itens.find((it) => it.id === id);
    if (!d) return;
    diligenciaForm.tipo = d.tipo;
    diligenciaForm.subcategorias = [...d.subcategorias];
    diligenciaForm.detalhe = d.detalhe;
    editingId = id;
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

  function itensToPayload(itemList: typeof itens): ItemPayload[] {
    return itemList.map((it) => ({
      tipo: it.tipo,
      subcategorias: it.subcategorias,
      detalhe: it.detalhe,
    }));
  }

  function errorMessage(err: unknown, fallback: string): string {
    if (err instanceof Error && err.message) {
      return err.message;
    }
    return fallback;
  }

  async function salvarItens(novaLista: ItemPayload[]): Promise<boolean> {
    try {
      await salvarDiligencia({ paId: data.processo.id, itens: novaLista });
      await invalidateAll();
      return true;
    } catch (err) {
      toast.error(errorMessage(err, "Não foi possível salvar as diligências"));
      return false;
    }
  }

  async function handleAddOrEdit(e: SubmitEvent) {
    e.preventDefault();
    if (isSubmitting) return;

    if (categoriaAtual?.subcategorias) {
      if (diligenciaForm.subcategorias.length === 0) {
        toast.error("Selecione ao menos um documento aplicável");
        return;
      }
    } else if (diligenciaForm.detalhe.trim() === "") {
      toast.error("Informe o detalhe da diligência");
      return;
    }

    const novoItem: ItemPayload = {
      tipo: diligenciaForm.tipo,
      subcategorias: [...diligenciaForm.subcategorias],
      detalhe: diligenciaForm.detalhe,
    };

    const base = itensToPayload(itens);
    let novaLista: ItemPayload[];
    if (editingId !== null) {
      const idx = itens.findIndex((it) => it.id === editingId);
      if (idx === -1) {
        novaLista = [...base, novoItem];
      } else {
        novaLista = base.map((it, i) => (i === idx ? novoItem : it));
      }
    } else {
      novaLista = [...base, novoItem];
    }

    isSubmitting = true;
    const ok = await salvarItens(novaLista);
    isSubmitting = false;

    if (ok) {
      open = false;
      resetDiligenciaForm();
    }
  }

  async function handleRemove(id: number) {
    if (isSubmitting) return;
    const novaLista = itensToPayload(itens.filter((it) => it.id !== id));
    isSubmitting = true;
    const ok = await salvarItens(novaLista);
    isSubmitting = false;
    if (ok) {
      toast.success("Diligência removida");
    }
  }

  async function handleEnviar() {
    if (isSubmitting) return;
    isSubmitting = true;
    try {
      await enviarDiligencia({ paId: data.processo.id });
      toast.success("Diligência registrada");
      await invalidateAll();
      await goto("/analise");
    } catch (err) {
      toast.error(errorMessage(err, "Não foi possível registrar a diligência"));
    } finally {
      isSubmitting = false;
    }
  }

  async function handleDescartar() {
    if (isSubmitting) return;
    isSubmitting = true;
    try {
      await descartarDiligencia({ paId: data.processo.id });
      toast.success("Rascunho descartado");
      await invalidateAll();
    } catch (err) {
      toast.error(errorMessage(err, "Não foi possível descartar o rascunho"));
    } finally {
      isSubmitting = false;
    }
  }
</script>

<svelte:head>
  <title>Registrar Diligência - Fila Aposentadoria</title>
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
        buttonText="Adicionar"
        buttonVariant="outline"
        bind:open
        onOpenChange={(isOpen) => {
          if (!isOpen) resetDiligenciaForm();
        }}
      >
        {#snippet buttonIcon()}
          <PlusIcon />
        {/snippet}
        {#snippet title()}
          {editingId !== null ? "Editar Diligência" : "Adicionar Diligência"}
        {/snippet}
        {#snippet description()}
          Preencha o tipo e informações da diligência
        {/snippet}

        <form class="flex flex-col gap-4 mt-6" onsubmit={handleAddOrEdit}>
          <div class="grid gap-1 min-w-0">
            <Label for="diligencia-tipo">Tipo de diligência</Label>
            <Select
              id="diligencia-tipo"
              required
              class="w-full min-w-0"
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
          </div>

          {#if categoriaAtual?.subcategorias}
            <fieldset class="space-y-2">
              <legend class="text-sm font-medium">
                Documentos aplicáveis
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
            <div class="grid gap-1">
              <Label for="diligencia-detalhe">Detalhe</Label>
              <Textarea
                id="diligencia-detalhe"
                bind:value={diligenciaForm.detalhe}
                placeholder="Detalhar a diligência"
                rows={5}
                required
              ></Textarea>
            </div>
          {/if}

          <div class="flex justify-end">
            <Button disabled={isSubmitting}>
              {editingId !== null ? "Salvar" : "Adicionar"}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>

    {#if itens.length > 0}
      <ul class="space-y-2">
        {#each itens as diligencia, i (diligencia.id)}
          <li
            class="flex items-center gap-3 rounded-lg border border-border p-3"
          >
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium">
                {i + 1}. {diligencia.tipo}
              </p>
              {#if diligencia.subcategorias.length > 0}
                <Popover.Root>
                  <Popover.Trigger
                    class="mt-0.5 text-xs text-muted-foreground underline cursor-pointer"
                  >
                    {diligencia.subcategorias.length}
                    {diligencia.subcategorias.length === 1
                      ? "documento selecionado"
                      : "documentos selecionados"}
                  </Popover.Trigger>
                  <Popover.Portal>
                    <Popover.Content
                      class="border border-border-strong shadow-sm z-30 max-w-80 rounded-xl p-4 w-full bg-surface space-y-1"
                      sideOffset={8}
                    >
                      <p class="text-sm font-medium">
                        Documentos selecionados:
                      </p>
                      <ul
                        class="text-sm text-muted-foreground space-y-0.5 list-disc max-h-60 overflow-y-auto"
                      >
                        {#each diligencia.subcategorias as sub}
                          <li class="ml-6">
                            <p class="tracking-tight">{sub}</p>
                          </li>
                        {/each}
                      </ul>
                    </Popover.Content>
                  </Popover.Portal>
                </Popover.Root>
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
                    onclick={() => editDiligencia(diligencia.id)}
                    disabled={isSubmitting}
                    class="text-primary bg-surface-alt hover:bg-surface-alt/80 cursor-pointer rounded-md p-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
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
                    onclick={() => handleRemove(diligencia.id)}
                    disabled={isSubmitting}
                    class="text-destructive bg-surface-alt hover:bg-surface-alt/80 cursor-pointer rounded-md p-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
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
      <div class="flex flex-col items-center py-10 border border-border">
        <p class="text-center font-semibold">Nenhuma diligência adicionada.</p>
        <p class="text-center text-sm text-muted-foreground">
          Utilize o botão acima para adicionar diligências.
        </p>
      </div>
    {/if}

    <div class="flex justify-end gap-2">
      {#if itens.length > 0}
        <AlertDialog
          buttonText="Descartar Rascunho"
          variant="destructive"
          disabled={isSubmitting}
          onConfirmed={handleDescartar}
        >
          {#snippet title()}
            Descartar rascunho
          {/snippet}
          {#snippet description()}
            Todas as diligências adicionadas serão removidas. Esta ação não pode
            ser desfeita.
          {/snippet}
        </AlertDialog>
      {/if}

      <AlertDialog
        buttonText="Registrar Diligência"
        disabled={itens.length === 0 || isSubmitting}
        onConfirmed={handleEnviar}
      >
        {#snippet title()}
          Confirmar Registro de Diligência
        {/snippet}
        {#snippet description()}
          Revise as diligências antes de registrar.
        {/snippet}

        <ul class="mt-4 max-h-72 overflow-y-auto space-y-3 text-sm">
          {#each itens as diligencia, i (diligencia.id)}
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
