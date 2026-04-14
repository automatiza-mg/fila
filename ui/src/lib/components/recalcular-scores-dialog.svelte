<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { recalcularScoresCmd } from "../../routes/(gestor)/processos/processo.remote";
  import { toast } from "svelte-sonner";
  import ArrowsClockwiseIcon from "phosphor-svelte/lib/ArrowsClockwiseIcon";
  import AlertDialog from "./ui/alert-dialog.svelte";

  let open = $state(false);

  async function handleConfirm() {
    try {
      await recalcularScoresCmd(null);
      open = false;
      toast.success("Recálculo de scores enfileirado com sucesso");
      await invalidateAll();
    } catch {
      toast.error(
        "Não foi possível recalcular os scores, tente novamente mais tarde",
      );
    }
  }
</script>

<AlertDialog
  buttonText="Recalcular Scores"
  variant="outline"
  bind:open
  onConfirmed={handleConfirm}
>
  {#snippet buttonIcon()}<ArrowsClockwiseIcon />{/snippet}
  {#snippet title()}Recalcular Scores{/snippet}
  {#snippet description()}
    Os scores de todos os processos serão recalculados. Esta operação será
    processada em segundo plano.
  {/snippet}
</AlertDialog>
