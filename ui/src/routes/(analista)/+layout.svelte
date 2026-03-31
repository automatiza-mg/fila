<script lang="ts">
  import { page } from "$app/state";
  import { cn } from "$lib/cn";
  import { sairForm } from "../(auth)/auth.remote";

  let { children } = $props();

  function isActive(href: string): boolean {
    return (
      page.url.pathname === href || page.url.pathname.startsWith(href + "/")
    );
  }
</script>

{#snippet tab(href: string, label: string)}
  <a
    {href}
    class={cn(
      "min-w-40 rounded-t-2xl p-3 text-center font-semibold",
      isActive(href) ? "bg-muted" : "bg-muted/60",
    )}
  >
    {label}
  </a>
{/snippet}

<div class="from-secondary to-primary flex min-h-svh flex-col bg-linear-to-b">
  <header class="bg-surface flex p-6 justify-between items-center">
    <div>
      <p class="text-xl font-bold">Fila Aposentadoria</p>
    </div>
    <div>
      <form {...sairForm}>
        <button class="font-semibold">Sair</button>
      </form>
    </div>
  </header>

  <main class="flex grow flex-col p-1.5 sm:p-4">
    <section class="mx-auto flex w-full max-w-7xl grow flex-col p-1.5 sm:p-4">
      <nav class="flex gap-0.5">
        {@render tab("/analista", "Análise")}
      </nav>

      <div
        class="bg-muted flex h-full grow flex-col rounded-tr-2xl rounded-b-2xl p-4 sm:p-8"
      >
        <div
          class="grow overflow-y-auto rounded-xl bg-surface p-6 shadow border border-secondary"
        >
          {@render children()}
        </div>
      </div>
    </section>
  </main>
</div>
