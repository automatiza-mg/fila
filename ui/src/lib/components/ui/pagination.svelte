<script lang="ts">
  import CaretLeftIcon from "phosphor-svelte/lib/CaretLeftIcon";
  import CaretRightIcon from "phosphor-svelte/lib/CaretRightIcon";
  import type { Paginated } from "$lib/api/types";

  type Props = {
    data: Paginated<unknown>;
    url: URL;
  };

  let { data, url }: Props = $props();

  function pageHref(page: number): string {
    const next = new URL(url);
    next.searchParams.set("page", String(page));
    return `${next.pathname}${next.search}`;
  }
</script>

{#if data.total_pages > 1}
  <nav class="mt-auto flex items-center justify-between border-t border-border pt-4">
    <a
      href={data.has_prev ? pageHref(data.current_page - 1) : undefined}
      class="inline-flex items-center gap-1 rounded-lg border border-border-strong px-3 py-1.5 text-sm font-medium outline-none focus-visible:ring-3 focus-visible:ring-secondary/50 focus-visible:border-secondary {data.has_prev
        ? 'hover:bg-surface-alt'
        : 'pointer-events-none opacity-40'}"
      aria-disabled={!data.has_prev}
      tabindex={data.has_prev ? 0 : -1}
    >
      <CaretLeftIcon class="size-4" />
      Anterior
    </a>

    <span class="text-sm text-muted-foreground">
      Página {data.current_page} de {data.total_pages}
    </span>

    <a
      href={data.has_next ? pageHref(data.current_page + 1) : undefined}
      class="inline-flex items-center gap-1 rounded-lg border border-border-strong px-3 py-1.5 text-sm font-medium outline-none focus-visible:ring-3 focus-visible:ring-secondary/50 focus-visible:border-secondary {data.has_next
        ? 'hover:bg-surface-alt'
        : 'pointer-events-none opacity-40'}"
      aria-disabled={!data.has_next}
      tabindex={data.has_next ? 0 : -1}
    >
      Próxima
      <CaretRightIcon class="size-4" />
    </a>
  </nav>
{/if}
