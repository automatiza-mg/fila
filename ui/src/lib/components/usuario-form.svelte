<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { createUsuarioForm } from "$lib/fns/usuarios.remote";
  import { formatCpf } from "$lib/formatter";
  import { toast } from "svelte-sonner";
  import Alert from "./ui/alert.svelte";
  import Button from "./ui/button.svelte";
  import Dialog from "./ui/dialog.svelte";
  import FormField from "./ui/form-field.svelte";
  import Input from "./ui/input.svelte";
  import Select from "./ui/select.svelte";

  let open = $state(false);

  function getOpen() {
    return open;
  }

  function setOpen(newOpen: boolean) {
    open = newOpen;
  }

  $effect(() => {
    const value = createUsuarioForm.fields.cpf.value() ?? "";
    createUsuarioForm.fields.cpf.set(formatCpf(value));
  });
</script>

<Dialog bind:open={getOpen, setOpen} buttonText="Novo Cadastro">
  {#snippet title()}
    Cadastrar Usuário
  {/snippet}

  {#snippet description()}
    Preencha os dados abaixo para cadastrar um novo usuário no sistema.
  {/snippet}

  <div class="pt-6">
    <form
      {...createUsuarioForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          form.reset();
          toast.success("Usuário cadastrado com sucesso!");
          setOpen(false);
          await invalidateAll();
        } catch {
          toast.error(
            "Não foi possível criar o usuário, tente novamente mais tarde",
          );
        }
      })}
      class="flex flex-col gap-4"
    >
      {#each createUsuarioForm.fields.issues() as issue}
        <Alert message={issue.message} variant="danger" />
      {/each}

      <FormField label="Nome" id="nome">
        <Input
          {...createUsuarioForm.fields.nome.as("text")}
          id="nome"
          required
        />
      </FormField>

      <FormField
        label="CPF"
        id="cpf"
        issues={createUsuarioForm.fields.cpf.issues()}
      >
        <Input
          {...createUsuarioForm.fields.cpf.as("text")}
          id="cpf"
          required
        />
      </FormField>

      <FormField
        label="Email"
        id="email"
        issues={createUsuarioForm.fields.email.issues()}
      >
        <Input
          {...createUsuarioForm.fields.email.as("email")}
          id="email"
          required
        />
      </FormField>

      <FormField
        label="Papel"
        id="papel"
        issues={createUsuarioForm.fields.papel.issues()}
      >
        <Select
          {...createUsuarioForm.fields.papel.as("select")}
          id="papel"
          required
        >
          <option value="ANALISTA">Analista</option>
          <option value="GESTOR">Gestor(a)</option>
          <option value="SUBSECRETARIO">Subsecretário(a)</option>
        </Select>
      </FormField>

      <div class="flex justify-end">
        <Button>Enviar</Button>
      </div>
    </form>
  </div>
</Dialog>
