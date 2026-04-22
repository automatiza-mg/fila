<script lang="ts">
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { toast } from "svelte-sonner";
  import { cadastrarForm } from "$lib/fns/auth.remote";

  let { data } = $props();

  $effect(() => {
    const issues = cadastrarForm.fields.issues();
    if (issues) {
      toast.error(issues[0].message);
    }
  });
</script>

<svelte:head>
  <title>Cadastrar - Fila Aposentadoria</title>
</svelte:head>

<h1 class="text-3xl font-bold text-center">Concluir Cadastro</h1>

<p class="text-sm text-muted-foreground">
  Olá, <span class="font-semibold">{data.usuario.nome}</span>. Defina sua senha
  abaixo para concluir o cadastro.
</p>

<form
  class="flex flex-col gap-8"
  {...cadastrarForm.enhance(async ({ submit }) => {
    try {
      await submit();
    } catch (err) {
      console.log(err);
    }
  })}
>
  <input type="hidden" name="token" value={data.token} />
  <input type="hidden" name="cpf" value={data.usuario.cpf} />

  <div class="flex flex-col gap-4">
    <FormField
      label="Senha"
      id="senha"
      issues={cadastrarForm.fields._senha.issues()}
    >
      <Input
        id="senha"
        {...cadastrarForm.fields._senha.as("password")}
        required
      />
    </FormField>

    <FormField
      label="Confirmar Senha"
      id="confirmar_senha"
      issues={cadastrarForm.fields._confirmar_senha.issues()}
    >
      <Input
        id="confirmar_senha"
        {...cadastrarForm.fields._confirmar_senha.as("password")}
        required
      />
    </FormField>
  </div>

  <Button>Cadastrar</Button>
</form>
