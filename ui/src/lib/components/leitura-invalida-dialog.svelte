<script lang="ts">
  import Alert from "$lib/components/ui/alert.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Dialog from "$lib/components/ui/dialog.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Textarea from "$lib/components/ui/textarea.svelte";
  import { leituraInvalidaForm } from "../../routes/(analista)/analista/analista.remote";
  import { toast } from "svelte-sonner";
  import { invalidateAll } from "$app/navigation";

  type Props = {
    processoId: number;
  };

  let { processoId }: Props = $props();

  let dialogOpen = $state(false);

  function getDialogOpen() {
    return dialogOpen;
  }

  function setDialogOpen(newOpen: boolean) {
    dialogOpen = newOpen;
  }
</script>

<Dialog
  bind:open={getDialogOpen, setDialogOpen}
  buttonText="Não é processo de aposentadoria"
  buttonVariant="destructive"
>
  {#snippet title()}
    Marcar como Leitura Inválida
  {/snippet}

  {#snippet description()}
    O processo será marcado como leitura inválida e desatribuído. Informe o
    motivo abaixo.
  {/snippet}

  <div class="pt-4">
    <form
      {...leituraInvalidaForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          form.reset();
          toast.success("Processo marcado como leitura inválida");
          setDialogOpen(false);
          await invalidateAll();
        } catch {
          toast.error(
            "Não foi possível marcar o processo como leitura inválida",
          );
        }
      })}
      class="flex flex-col gap-4"
    >
      {#each leituraInvalidaForm.fields.issues() as issue}
        <Alert message={issue.message} variant="danger" />
      {/each}

      <input type="hidden" name="processoId" value={processoId} />

      <FormField
        label="Motivo"
        id="motivo"
        issues={leituraInvalidaForm.fields._motivo.issues()}
      >
        <Textarea
          {...leituraInvalidaForm.fields._motivo.as("text")}
          id="motivo"
          placeholder="Descreva o motivo..."
          rows={4}
          required
        />
      </FormField>

      <div class="flex justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onclick={() => setDialogOpen(false)}
        >
          Cancelar
        </Button>
        <Button>Confirmar</Button>
      </div>
    </form>
  </div>
</Dialog>
