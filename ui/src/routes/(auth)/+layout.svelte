<script lang="ts">
  let { children } = $props();
</script>

<main
  class="bg-linear-to-b from-secondary flex flex-col items-center justify-center p-4 to-primary min-h-svh"
>
  <section
    class="max-w-100 w-full bg-white px-6 py-8 rounded-4xl shadow-sm space-y-4"
  >
    {@render children()}
  </section>
</main>
