<script lang="ts">
  import Pendencias from "$lib/components/pendencias.svelte";
  import UsuarioForm from "$lib/components/usuario-form.svelte";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Usuários | Fila Aposentadoria</title>
</svelte:head>

<div class="space-y-4">
  <div class="flex">
    <UsuarioForm />
  </div>

  <div>
    <table class="w-full border border-border text-sm">
      <thead>
        <tr class="border-y border-border bg-surface-alt">
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
                  papel={usuario.papel}
                />
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
