<script lang="ts">
  let { children } = $props();
</script>

<main
  class="flex flex-col items-center justify-center min-h-svh p-4 bg-linear-to-b from-escritorio-light to-escritorio"
>
  <section
    class="max-w-105 w-full px-8 py-10 space-y-6 bg-white rounded-4xl shadow-md"
  >
    {@render children()}
  </section>
</main>
