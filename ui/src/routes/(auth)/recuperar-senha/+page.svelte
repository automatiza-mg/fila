<script lang="ts">
  import Button from "$lib/components/ui/button.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { formatCpf } from "$lib/formatter";
  import { toast } from "svelte-sonner";
  import { recuperarSenhaForm } from "../auth.remote";

  let enviado = $state(false);

  $effect(() => {
    const value = recuperarSenhaForm.fields.cpf.value() ?? "";
    recuperarSenhaForm.fields.cpf.set(formatCpf(value));
  });

  $effect(() => {
    const issues = recuperarSenhaForm.fields.issues();
    if (issues) {
      toast.error(issues[0].message);
    }
  });
</script>

<svelte:head>
  <title>Recuperar Senha - Fila Aposentadoria</title>
</svelte:head>

<h1 class="text-3xl font-bold text-center">Recuperar Senha</h1>

{#if enviado}
  <p class="text-center text-muted-foreground">
    Se o CPF informado estiver cadastrado, você receberá um e-mail com
    instruções para redefinir sua senha.
  </p>

  <Button href="/entrar" variant="outline">Voltar para login</Button>
{:else}
  <p class="text-center text-muted-foreground">
    Informe seu CPF para receber um e-mail de recuperação de senha.
  </p>

  <form
    class="flex flex-col gap-8"
    {...recuperarSenhaForm.enhance(async ({ submit }) => {
      try {
        await submit();
        enviado = true;
      } catch (err) {
        console.log(err);
      }
    })}
  >
    <div class="flex flex-col gap-4">
      <FormField
        label="CPF"
        id="cpf"
        issues={recuperarSenhaForm.fields.cpf.issues()}
      >
        <Input
          id="cpf"
          {...recuperarSenhaForm.fields.cpf.as("text")}
          required
        />
      </FormField>
    </div>

    <div class="flex flex-col gap-2">
      <Button>Enviar</Button>
      <Button href="/entrar" variant="outline">Voltar para login</Button>
    </div>
  </form>
{/if}
