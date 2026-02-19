<script lang="ts">
  import { login } from "./login.remote";
  import Eye from "@lucide/svelte/icons/eye";
  import EyeOff from "@lucide/svelte/icons/eye-off";

  let showPassword = $state(false);
  let cpf = $state("");

  function togglePassword() {
    showPassword = !showPassword;
  }

  $effect(() => {
    // Remove non-digits, limit to 11 digits, and format as XXX.XXX.XXX-XX
    cpf = cpf
      .replace(/\D/g, "")
      .slice(0, 11)
      .replace(/(\d{3})(\d)/, "$1.$2")
      .replace(/(\d{3})\.(\d{3})(\d)/, "$1.$2.$3")
      .replace(/(\d{3})\.(\d{3})\.(\d{3})(\d)/, "$1.$2.$3-$4");
  });
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
      <label for="cpf" class="font-medium">CPF</label>
      <input
        id="cpf"
        {...login.fields.cpf.as("text")}
        autocomplete="username"
        class="py-2 px-3 border rounded-2xl border-stone-200"
        bind:value={cpf}
        required
      />
      {#each login.fields.cpf.issues() as issue}
        <p>{issue.message}</p>
      {/each}
    </div>

    <div class="grid gap-1">
      <label for="senha" class="font-medium">Senha</label>

      <div class="relative">
        <input
          id="senha"
          {...login.fields._senha.as(showPassword ? "text" : "password")}
          autocomplete="current-password"
          class="py-2 px-3 border rounded-2xl border-stone-200 w-full"
          required
        />
        <button
          type="button"
          class="absolute top-1/2 right-3 transform -translate-y-1/2 text-stone-600"
          onclick={togglePassword}
          aria-label={showPassword ? "Esconder Senha" : "Mostrar Senha"}
        >
          {#if showPassword}
            <EyeOff />
          {:else}
            <Eye />
          {/if}
        </button>
      </div>

      <a href="/recuperar-senha" class="underline text-stone-600 w-fit">
        Esqueci minha senha
      </a>
      {#each login.fields._senha.issues() as issue}
        <p>{issue.message}</p>
      {/each}
    </div>
  </div>

  <button class="px-4 py-2 bg-escritorio text-white rounded-xl font-semibold">
    Enviar
  </button>
</form>
