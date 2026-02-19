<script lang="ts">
  import { login } from "./login.remote";

  let showPassword = $state(false);
  let cpf = $state("");

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

<main
  class="flex flex-col items-center justify-center min-h-svh p-4 bg-stone-100"
>
  <section class="max-w-md w-full p-8 space-y-6 bg-white rounded-4xl shadow-md">
    <div>
      <h1 class="text-center text-2xl font-bold uppercase">Entrar</h1>
    </div>

    {#each login.fields.issues() as issue}
      <div>
        {issue.message}
      </div>
    {/each}

    <form {...login} class="flex flex-col gap-4">
      <div class="grid gap-1">
        <label for="cpf" class="font-medium">CPF</label>
        <input
          id="cpf"
          {...login.fields.cpf.as("text")}
          autocomplete="username"
          class="p-2 border rounded-2xl border-stone-200"
          bind:value={cpf}
        />
        {#each login.fields.cpf.issues() as issue}
          <p>{issue.message}</p>
        {/each}
      </div>

      <div class="grid gap-1">
        <label for="senha" class="font-medium">Senha</label>
        <input
          id="senha"
          {...login.fields._senha.as(showPassword ? "text" : "password")}
          autocomplete="current-password"
          class="p-2 border rounded-2xl border-stone-200"
        />
        {#each login.fields._senha.issues() as issue}
          <p>{issue.message}</p>
        {/each}
      </div>

      <button class="px-4 py-2 bg-blue-900 text-white rounded-xl font-semibold"
        >Enviar</button
      >
    </form>
  </section>
</main>
