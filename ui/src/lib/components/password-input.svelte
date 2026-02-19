<script lang="ts">
  import Eye from "@lucide/svelte/icons/eye";
  import EyeOff from "@lucide/svelte/icons/eye-off";
  import type { HTMLInputAttributes } from "svelte/elements";

  let showPassword = $state(false);
  let { value = $bindable(), ...props }: HTMLInputAttributes = $props();

  function togglePassword() {
    showPassword = !showPassword;
  }
</script>

<div class="relative">
  <input
    class="py-2 px-3 border rounded-2xl border-stone-200 w-full"
    {...props}
    bind:value
    type={showPassword ? "text" : "password"}
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
