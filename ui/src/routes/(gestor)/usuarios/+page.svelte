<script lang="ts">
  import Pendencias from "$lib/components/pendencias.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Usuários | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-4">
  <div class="flex">
    <a
      href="/usuarios/criar"
      class="px-4 py-2 font-semibold bg-primary text-white rounded-xl border border-transparent"
    >
      Novo Cadastro
    </a>
  </div>

  <div>
    <table class="w-full border border-stone-200 text-sm">
      <thead>
        <tr class="border-y border-stone-200 bg-stone-100">
          <th class="text-left font-semibold p-2.5">Nome</th>
          <th class="text-left font-semibold p-2.5">CPF</th>
          <th class="text-left font-semibold p-2.5">Email</th>
          <th class="text-left font-semibold p-2.5">Papel</th>
          <th class="text-left font-semibold p-2.5">Pendências</th>
        </tr>
      </thead>
      <tbody>
        {#each data.usuarios as usuario}
          <tr>
            <td class="p-2.5">
              <a
                href={`/usuarios/${usuario.id}`}
                class="text-primary underline"
              >
                {usuario.nome}
              </a>
            </td>
            <td class="p-2.5">{usuario.cpf}</td>
            <td class="p-2.5">{usuario.email}</td>
            <td class="p-2.5">{usuario.papel ?? "Não possui"}</td>
            <td>
              {#if usuario.pendencias.length === 0}
                <span>Não possui</span>
              {:else}
                <Pendencias
                  usuarioId={usuario.id}
                  pendencias={usuario.pendencias}
                />
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
