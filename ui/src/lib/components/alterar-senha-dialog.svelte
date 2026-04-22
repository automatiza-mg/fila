<script lang="ts">
  import Alert from "$lib/components/ui/alert.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import Dialog from "$lib/components/ui/dialog.svelte";
  import FormField from "$lib/components/ui/form-field.svelte";
  import Input from "$lib/components/ui/input.svelte";
  import { alterarSenhaForm } from "$lib/fns/auth.remote";
  import KeyIcon from "phosphor-svelte/lib/KeyIcon";
  import { toast } from "svelte-sonner";

  let dialogOpen = $state(false);

  function getDialogOpen() {
    return dialogOpen;
  }

  function setDialogOpen(newOpen: boolean) {
    dialogOpen = newOpen;
  }
</script>

<Dialog
  bind:open={getDialogOpen, setDialogOpen}
  buttonText="Alterar senha"
  buttonVariant="outline"
>
  {#snippet buttonIcon()}
    <KeyIcon />
  {/snippet}

  {#snippet title()}
    Alterar senha
  {/snippet}

  {#snippet description()}
    Informe sua senha atual e defina uma nova senha.
  {/snippet}

  <div class="pt-4">
    <form
      {...alterarSenhaForm.enhance(async ({ form, submit }) => {
        try {
          await submit();
          const issues = alterarSenhaForm.fields.issues();
          if (issues && issues.length > 0) return;
          form.reset();
          toast.success("Senha alterada com sucesso");
          setDialogOpen(false);
        } catch {
          toast.error("Não foi possível alterar a senha");
        }
      })}
      class="flex flex-col gap-4"
    >
      {#each alterarSenhaForm.fields.issues() ?? [] as issue}
        <Alert message={issue.message} variant="danger" />
      {/each}

      <FormField
        label="Senha atual"
        id="senha_atual"
        issues={alterarSenhaForm.fields._senha_atual.issues()}
      >
        <Input
          {...alterarSenhaForm.fields._senha_atual.as("password")}
          id="senha_atual"
          autocomplete="current-password"
          required
        />
      </FormField>

      <FormField
        label="Nova senha"
        id="nova_senha"
        issues={alterarSenhaForm.fields._nova_senha.issues()}
      >
        <Input
          {...alterarSenhaForm.fields._nova_senha.as("password")}
          id="nova_senha"
          autocomplete="new-password"
          required
        />
      </FormField>

      <FormField
        label="Confirmar nova senha"
        id="confirmar_nova_senha"
        issues={alterarSenhaForm.fields._confirmar_nova_senha.issues()}
      >
        <Input
          {...alterarSenhaForm.fields._confirmar_nova_senha.as("password")}
          id="confirmar_nova_senha"
          autocomplete="new-password"
          required
        />
      </FormField>

      <div class="flex justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onclick={() => setDialogOpen(false)}
        >
          Cancelar
        </Button>
        <Button>Confirmar</Button>
      </div>
    </form>
  </div>
</Dialog>
