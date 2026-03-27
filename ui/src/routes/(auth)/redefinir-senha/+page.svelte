<script lang="ts">
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { toast } from "svelte-sonner";
  import { redefinirSenhaForm } from "../auth.remote";

  let { data } = $props();

  $effect(() => {
    const issues = redefinirSenhaForm.fields.issues();
    if (issues) {
      toast.error(issues[0].message);
    }
  });
</script>

<svelte:head>
  <title>Redefinir Senha - Fila Aposentadoria</title>
</svelte:head>

<h1 class="text-3xl font-bold text-center">Redefinir Senha</h1>

<p class="text-center text-muted-foreground">
  Olá, {data.usuario.nome}. Defina sua nova senha abaixo.
</p>

<form
  class="flex flex-col gap-8"
  {...redefinirSenhaForm.enhance(async ({ submit }) => {
    try {
      await submit();
    } catch (err) {
      console.log(err);
    }
  })}
>
  <input type="hidden" name="token" value={data.token} />

  <div class="flex flex-col gap-4">
    <FormField
      label="Nova Senha"
      id="senha"
      issues={redefinirSenhaForm.fields._senha.issues()}
    >
      <Input
        id="senha"
        {...redefinirSenhaForm.fields._senha.as("password")}
        required
      />
    </FormField>

    <FormField
      label="Confirmar Senha"
      id="confirmar_senha"
      issues={redefinirSenhaForm.fields._confirmar_senha.issues()}
    >
      <Input
        id="confirmar_senha"
        {...redefinirSenhaForm.fields._confirmar_senha.as("password")}
        required
      />
    </FormField>
  </div>

  <Button>Redefinir Senha</Button>
</form>
