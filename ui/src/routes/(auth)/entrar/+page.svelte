<script lang="ts">
  import Alert from "$lib/components/ui/alert.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { formatCpf } from "$lib/formatter";
  import { toast } from "svelte-sonner";
  import { entrarForm } from "../auth.remote";

  $effect(() => {
    const value = entrarForm.fields.cpf.value() ?? "";
    entrarForm.fields.cpf.set(formatCpf(value));
  });

  $effect(() => {
    const issues = entrarForm.fields.issues();
    if (issues) {
      toast.error(issues[0].message);
    }
  });
</script>

<svelte:head>
  <title>Entrar - Fila Aposentadoria</title>
</svelte:head>

<h1 class="text-3xl font-bold text-center">Entrar</h1>

<form class="flex flex-col gap-6" {...entrarForm}>
  <div class="flex flex-col gap-4">
    <FormField label="CPF" id="cpf" issues={entrarForm.fields.cpf.issues()}>
      <Input id="cpf" {...entrarForm.fields.cpf.as("text")} required />
    </FormField>

    <div class="space-y-1">
      <FormField
        label="Senha"
        id="senha"
        issues={entrarForm.fields._senha.issues()}
      >
        <Input
          id="senha"
          {...entrarForm.fields._senha.as("password")}
          required
        />
      </FormField>

      <div class="flex justify-end">
        <a
          href="/recuperar-senha"
          class="text-muted-foreground underline text-sm"
        >
          Esqueci minha senha
        </a>
      </div>
    </div>
  </div>

  <Button>Enviar</Button>
</form>
