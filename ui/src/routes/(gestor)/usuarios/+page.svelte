<script lang="ts">
  import Input from "$lib/components/input.svelte";
  import type { FormEventHandler } from "svelte/elements";
  import type { PageProps } from "./$types";
  import { criar } from "./criar.remote";

  let { data }: PageProps = $props();

  let dialogOpen = $state(false);

  const papelLabels: Record<string, string> = {
    ADMIN: "Administrador",
    ANALISTA: "Analista",
    GESTOR: "Gestor",
    SUBSECRETARIO: "Subsecretário",
  };

  const papelColors: Record<string, string> = {
    ADMIN: "bg-red-100 text-red-800",
    ANALISTA: "bg-blue-100 text-blue-800",
    GESTOR: "bg-purple-100 text-purple-800",
    SUBSECRETARIO: "bg-amber-100 text-amber-800",
  };

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

  let dialogEl: HTMLDialogElement | undefined = $state();

  function openDialog() {
    dialogOpen = true;
    dialogEl?.showModal();
  }

  function closeDialog() {
    dialogOpen = false;
    dialogEl?.close();
  }
</script>

<svelte:head>
  <title>Usuários | Fila Aposentadoria</title>
</svelte:head>

<div class="mb-6 flex items-center justify-between">
  <div>
    <h1 class="text-3xl font-bold text-escritorio">Usuários</h1>
    <p class="text-gray-600 mt-1">
      Total: {data.usuarios?.length ?? 0} usuários
    </p>
  </div>
  <button
    onclick={openDialog}
    class="px-4 py-2 bg-escritorio text-white rounded-xl font-semibold hover:opacity-90 transition-opacity"
  >
    Novo Usuário
  </button>
</div>

{#if data.usuarios && data.usuarios.length > 0}
  <div class="overflow-x-auto bg-white rounded-lg shadow">
    <table class="w-full">
      <thead>
        <tr class="border-b border-gray-200 bg-escritorio">
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >Nome</th
          >
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >CPF</th
          >
          <th class="px-6 py-3 text-left text-sm font-semibold text-white"
            >Email</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Email Verificado</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Papel</th
          >
          <th class="px-6 py-3 text-center text-sm font-semibold text-white"
            >Pendências</th
          >
        </tr>
      </thead>
      <tbody>
        {#each data.usuarios as usuario}
          <tr
            class="border-b border-gray-100 hover:bg-gray-50 transition-colors"
          >
            <td class="px-6 py-4 text-sm font-medium text-escritorio">
              <a href="/usuarios/{usuario.id}" class="hover:underline">
                {usuario.nome}
              </a>
            </td>
            <td class="px-6 py-4 text-sm text-gray-600 font-mono">
              {usuario.cpf}
            </td>
            <td class="px-6 py-4 text-sm text-gray-700">
              {usuario.email}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.email_verificado}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800"
                >
                  Sim
                </span>
              {:else}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-800"
                >
                  Não
                </span>
              {/if}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.papel}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold {papelColors[
                    usuario.papel
                  ] ?? 'bg-gray-100 text-gray-800'}"
                >
                  {papelLabels[usuario.papel] ?? usuario.papel}
                </span>
              {:else}
                <span class="text-gray-400">—</span>
              {/if}
            </td>
            <td class="px-6 py-4 text-center text-sm">
              {#if usuario.pendencias && usuario.pendencias.length > 0}
                <span
                  class="inline-block px-3 py-1 rounded-full text-xs font-semibold bg-yellow-100 text-yellow-800"
                >
                  {usuario.pendencias.length}
                </span>
              {:else}
                <span class="text-gray-400">—</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <div class="bg-white rounded-lg shadow p-12 text-center">
    <p class="text-gray-500 text-lg">Nenhum usuário encontrado.</p>
  </div>
{/if}

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<dialog
  bind:this={dialogEl}
  class="rounded-2xl shadow-xl p-0 w-full max-w-lg backdrop:bg-black/50"
  onclose={closeDialog}
  onkeydown={(e) => e.key === "Escape" && closeDialog()}
>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-bold text-escritorio">Novo Usuário</h2>
      <button
        type="button"
        onclick={closeDialog}
        class="text-gray-400 hover:text-gray-600 text-2xl leading-none"
      >
        &times;
      </button>
    </div>

    {#each criar.fields.issues() as issue}
      <div class="mb-4 p-3 bg-red-50 text-red-700 rounded-xl text-sm">
        {issue.message}
      </div>
    {/each}

    <form {...criar} class="flex flex-col gap-4">
      <div class="grid gap-1">
        <label for="nome" class="font-medium w-fit">Nome</label>
        <Input
          id="nome"
          required
          {...criar.fields.nome.as("text")}
        />
        {#each criar.fields.nome.issues() as issue}
          <p class="text-sm text-red-500">{issue.message}</p>
        {/each}
      </div>

      <div class="grid gap-1">
        <label for="cpf" class="font-medium w-fit">CPF</label>
        <Input
          id="cpf"
          required
          {...criar.fields.cpf.as("text")}
          oninput={formatCpf}
        />
        {#each criar.fields.cpf.issues() as issue}
          <p class="text-sm text-red-500">{issue.message}</p>
        {/each}
      </div>

      <div class="grid gap-1">
        <label for="email" class="font-medium w-fit">Email</label>
        <Input
          id="email"
          type="email"
          required
          {...criar.fields.email.as("text")}
        />
        {#each criar.fields.email.issues() as issue}
          <p class="text-sm text-red-500">{issue.message}</p>
        {/each}
      </div>

      <div class="grid gap-1">
        <label for="papel" class="font-medium w-fit">Papel</label>
        <select
          id="papel"
          required
          {...criar.fields.papel.as("select")}
          class="py-2 px-3 border rounded-2xl border-stone-200 w-full bg-white"
        >
          <option value="" disabled selected>Selecione um papel</option>
          <option value="ANALISTA">Analista</option>
          <option value="GESTOR">Gestor</option>
          <option value="SUBSECRETARIO">Subsecretário</option>
        </select>
        {#each criar.fields.papel.issues() as issue}
          <p class="text-sm text-red-500">{issue.message}</p>
        {/each}
      </div>

      <div class="flex gap-3 mt-2">
        <button
          type="button"
          onclick={closeDialog}
          class="flex-1 px-4 py-2 border border-stone-300 text-stone-700 rounded-xl font-semibold hover:bg-stone-50 transition-colors"
        >
          Cancelar
        </button>
        <button
          type="submit"
          class="flex-1 px-4 py-2 bg-escritorio text-white rounded-xl font-semibold hover:opacity-90 transition-opacity"
        >
          Criar Usuário
        </button>
      </div>
    </form>
  </div>
</dialog>
