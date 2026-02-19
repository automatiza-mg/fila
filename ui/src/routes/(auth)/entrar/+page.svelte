<script lang="ts">
  import Input from "$lib/components/input.svelte";
  import PasswordInput from "$lib/components/password-input.svelte";
  import type { FormEventHandler } from "svelte/elements";
  import { login } from "./login.remote";

  const formatCpf: FormEventHandler<HTMLInputElement> = (e) => {
    const target = e.currentTarget;
    let value = target.value;
    target.value = value
      .replace(/\D/g, "")
      .slice(0, 11)
      .replace(/(\d{3})(\d)/, "$1.$2")
      .replace(/(\d{3})\.(\d{3})(\d)/, "$1.$2.$3")
      .replace(/(\d{3})\.(\d{3})\.(\d{3})(\d)/, "$1.$2.$3-$4");
  };
</script>

<svelte:head>
  <title>Entrar | Fila Aposentadoria</title>
</svelte:head>

<div>
  <h1 class="text-center text-3xl font-bold">Entrar</h1>
</div>

{#each login.fields.issues() as issue}
  <div>
    {issue.message}
  </div>
{/each}

<form {...login} class="flex flex-col gap-8">
  <div class="space-y-4">
    <div class="grid gap-1">
      <label for="cpf" class="font-medium w-fit">CPF</label>
      <Input
        id="cpf"
        autocomplete="username"
        required
        {...login.fields.cpf.as("text")}
        oninput={formatCpf}
      />
      {#each login.fields.cpf.issues() as issue}
        <p class="text-sm text-red-500">{issue.message}</p>
      {/each}
    </div>

    <div class="grid gap-1">
      <label for="senha" class="font-medium w-fit">Senha</label>
      <PasswordInput
        id="senha"
        {...login.fields._senha.as("password")}
        required
        minlength={8}
        maxlength={60}
        autocomplete="current-password"
      />

      {#each login.fields._senha.issues() as issue}
        <p class="text-sm text-red-500">{issue.message}</p>
      {/each}

      <a href="/recuperar-senha" class="underline text-stone-600 w-fit">
        Esqueci minha senha
      </a>
    </div>
  </div>

  <button class="px-4 py-2 bg-escritorio text-white rounded-xl font-semibold">
    Enviar
  </button>
</form>
