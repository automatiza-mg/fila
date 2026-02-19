<script lang="ts">
  import Input from "$lib/components/input.svelte";
  import PasswordInput from "$lib/components/password-input.svelte";
  import type { PageProps } from "./$types";
  import { registrar } from "./register.remote";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Cadastrar | Fila Aposentadoria</title>
</svelte:head>

<form {...registrar} class="flex flex-col gap-4">
  <input {...registrar.fields.cpf.as("hidden", data.usuario.cpf)} />
  <input {...registrar.fields.token.as("hidden", data.token)} />

  <div class="grid gap-1">
    <label for="nome" class="font-medium w-fit">Nome</label>
    <Input id="nome" readonly value={data.usuario.nome} />
  </div>

  <div class="grid gap-1">
    <label for="senha" class="font-medium w-fit">Senha</label>
    <PasswordInput
      id="senha"
      required
      minlength={8}
      maxlength={60}
      {...registrar.fields._senha.as("password")}
    />
  </div>

  <div class="grid gap-1">
    <label for="confirmar_senha" class="font-medium w-fit">Repetir Senha</label>
    <PasswordInput
      id="confirmar_senha"
      required
      minlength={8}
      maxlength={60}
      {...registrar.fields._confirmar_senha.as("password")}
    />
  </div>

  <button class="px-4 py-2 bg-escritorio text-white rounded-xl font-semibold">
    Enviar
  </button>
</form>
