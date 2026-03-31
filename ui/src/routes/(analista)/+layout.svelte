<script lang="ts">
  import { sairForm } from "../(auth)/auth.remote";

  let { children } = $props();
</script>

<div class="from-secondary to-primary flex min-h-svh flex-col bg-linear-to-b">
  <header class="bg-white flex p-6 justify-between items-center">
    <div>
      <p class="text-xl font-bold">Fila Aposentadoria</p>
    </div>
    <div>
      <form {...sairForm}>
        <button class="font-semibold">Sair</button>
      </form>
    </div>
  </header>

  <main class="flex grow flex-col p-4">
    <section class="mx-auto flex w-full max-w-7xl grow flex-col p-4">
      <nav class="flex gap-px">
        <a
          href="/analista"
          class="bg-muted min-w-40 rounded-t-2xl p-3 text-center font-semibold"
        >
          Análise
        </a>
      </nav>

      <div
        class="bg-muted flex h-full grow flex-col rounded-tr-2xl rounded-b-2xl p-8"
      >
        <div
          class="grow overflow-y-auto rounded-xl bg-white p-6 shadow border border-secondary"
        >
          {@render children()}
        </div>
      </div>
    </section>
  </main>
</div>
