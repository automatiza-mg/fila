<script lang="ts">
  import Alert from "$lib/components/ui/alert.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import Select from "$lib/components/ui/select.svelte";
  import { formatCpf } from "$lib/formatter";
  import { createUsuarioForm } from "../usuario.remote";

  $effect(() => {
    const value = createUsuarioForm.fields.cpf.value() ?? "";
    createUsuarioForm.fields.cpf.set(formatCpf(value));
  });
</script>

<svelte:head>
  <title>Cadastrar Usuário | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-6">
  <h1 class="text-2xl font-semibold text-center">Cadastrar Usuário</h1>

  <form {...createUsuarioForm} class="flex flex-col gap-4 max-w-md mx-auto">
    {#each createUsuarioForm.fields.issues() as issue}
      <Alert message={issue.message} variant="danger" />
    {/each}

    <FormField label="Nome" id="nome">
      <Input {...createUsuarioForm.fields.nome.as("text")} id="nome" required />
    </FormField>

    <FormField
      label="CPF"
      id="cpf"
      issues={createUsuarioForm.fields.cpf.issues()}
    >
      <Input {...createUsuarioForm.fields.cpf.as("text")} id="cpf" required />
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

    <Button>Enviar</Button>
  </form>
</div>
