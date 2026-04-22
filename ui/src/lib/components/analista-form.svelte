<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import {
    criarAnalistaForm,
    unidadesQuery,
  } from "$lib/fns/usuarios.remote";
  import { toast } from "svelte-sonner";
  import Alert from "./ui/alert.svelte";
  import Button from "./ui/button.svelte";
  import Dialog from "./ui/dialog.svelte";
  import FormField from "./ui/form-field.svelte";
  import Select from "./ui/select.svelte";

  type Props = {
    usuarioId: number;
    buttonText?: string;
    buttonVariant?: "default" | "destructive" | "outline" | "link";
  };

  let {
    usuarioId,
    buttonText = "Cadastrar Analista",
    buttonVariant = "default",
  }: Props = $props();

  let open = $state(false);

  function getOpen() {
    return open;
  }

  function setOpen(newOpen: boolean) {
    open = newOpen;
  }

  const unidades = unidadesQuery();
</script>

<Dialog bind:open={getOpen, setOpen} {buttonText} {buttonVariant}>
  {#snippet title()}
    Cadastrar Analista
  {/snippet}

  {#snippet description()}
    Preencha os dados abaixo para cadastrar o analista.
  {/snippet}

  <div class="pt-6">
    <form
      {...criarAnalistaForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          form.reset();
          toast.success("Analista cadastrado com sucesso!");
          setOpen(false);
          await invalidateAll();
        } catch {
          toast.error(
            "Não foi possível cadastrar o analista, tente novamente mais tarde",
          );
        }
      })}
      class="flex flex-col gap-4"
    >
      {#each criarAnalistaForm.fields.issues() as issue}
        <Alert message={issue.message} variant="danger" />
      {/each}

      <input type="hidden" name="usuarioId" value={usuarioId} />

      <FormField
        label="Órgão de Exercício"
        id="orgao"
        issues={criarAnalistaForm.fields.orgao.issues()}
      >
        <Select
          {...criarAnalistaForm.fields.orgao.as("select")}
          id="orgao"
          required
        >
          <option value="">Selecione</option>
          <option value="SEPLAG">SEPLAG</option>
          <option value="SEE">SEE</option>
        </Select>
      </FormField>

      <FormField
        label="Caixa do SEI"
        id="sei_unidade_id"
        issues={criarAnalistaForm.fields.sei_unidade_id.issues()}
      >
        <Select
          {...criarAnalistaForm.fields.sei_unidade_id.as("select")}
          id="sei_unidade_id"
          required
          disabled={!unidades.ready}
        >
          <option value="">
            {unidades.ready ? "Selecione" : "Carregando..."}
          </option>
          {#if unidades.ready}
            {#each unidades.current as unidade}
              <option value={unidade.id}>{unidade.sigla}</option>
            {/each}
          {/if}
        </Select>
      </FormField>

      <div class="flex justify-end">
        <Button>Enviar</Button>
      </div>
    </form>
  </div>
</Dialog>
