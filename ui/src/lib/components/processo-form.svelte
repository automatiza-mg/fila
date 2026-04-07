<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { criarProcessoForm } from "../../routes/(gestor)/processos/processo.remote";
  import { toast } from "svelte-sonner";
  import Alert from "./ui/alert.svelte";
  import Button from "./ui/button.svelte";
  import Dialog from "./ui/dialog.svelte";
  import FormField from "./ui/form-field.svelte";
  import Input from "./ui/input.svelte";

  let open = $state(false);

  function getOpen() {
    return open;
  }

  function setOpen(newOpen: boolean) {
    open = newOpen;
  }
</script>

<Dialog bind:open={getOpen, setOpen} buttonText="Novo Processo">
  {#snippet title()}
    Criar Processo
  {/snippet}

  {#snippet description()}
    Informe o número do processo para cadastrá-lo no sistema.
  {/snippet}

  <div class="pt-6">
    <form
      {...criarProcessoForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          form.reset();
          toast.success("Processo criado com sucesso!");
          setOpen(false);
          await invalidateAll();
        } catch {
          toast.error(
            "Não foi possível criar o processo, tente novamente mais tarde",
          );
        }
      })}
      class="flex flex-col gap-4"
    >
      {#each criarProcessoForm.fields.issues() as issue}
        <Alert message={issue.message} variant="danger" />
      {/each}

      <FormField
        label="Número"
        id="numero"
        issues={criarProcessoForm.fields.numero.issues()}
      >
        <Input
          {...criarProcessoForm.fields.numero.as("text")}
          id="numero"
          placeholder="Número do processo"
          required
        />
      </FormField>

      <div class="flex justify-end">
        <Button>Enviar</Button>
      </div>
    </form>
  </div>
</Dialog>
