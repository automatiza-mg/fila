<script lang="ts">
  import { criarPrioridadeForm } from "../../routes/(gestor)/processos/[id]/processo.remote";
  import Dialog from "./ui/dialog.svelte";
  import FormField from "./ui/form-field.svelte";
  import Textarea from "./ui/textarea.svelte";
  import { toast } from "svelte-sonner";

  type Props = {
    paId: number;
  };

  let { paId }: Props = $props();
  let open = $state(false);

  function getOpen() {
    return open;
  }

  function setOpen(newOpen: boolean) {
    open = newOpen;
  }
</script>

<Dialog bind:open={getOpen, setOpen} buttonText="Solicitar Prioridade">
  {#snippet title()}
    Solicitar Prioridade
  {/snippet}

  {#snippet description()}
    Preencha uma justificativa para solicitar a priorização na análise do
    processo.
  {/snippet}

  <div class="pt-6">
    <form
      {...criarPrioridadeForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          form.reset();
          toast("Solicitação de prioridade criada", {
            description: "Os usuários responsáveis serão notificados em breve",
          });
          setOpen(false);
        } catch (err) {
          toast("Algo deu errado ao criar a solicitação");
        }
      })}
      class="flex flex-col gap-4"
    >
      <input
        {...criarPrioridadeForm.fields.paId.as("hidden", paId.toString())}
      />

      <FormField
        id="justificativa"
        label="Justificativa"
        issues={criarPrioridadeForm.fields.justificativa.issues()}
      >
        <Textarea
          id="justificativa"
          rows={6}
          required
          {...criarPrioridadeForm.fields.justificativa.as("text")}
        />
      </FormField>

      <div class="flex justify-end">
        <button
          class="px-4 py-2 font-semibold bg-primary text-white rounded-2xl border border-transparent"
        >
          Enviar
        </button>
      </div>
    </form>
  </div>
</Dialog>
