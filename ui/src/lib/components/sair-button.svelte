<script lang="ts">
  import { sairForm } from "$lib/fns/auth.remote";
  import SignOutIcon from "phosphor-svelte/lib/SignOutIcon";
  import Button from "./ui/button.svelte";
</script>

<form {...sairForm}>
  <Button variant="destructive" type="submit">
    <SignOutIcon />
    Sair
  </Button>
</form>
