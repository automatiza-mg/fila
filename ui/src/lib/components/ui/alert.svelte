<script lang="ts">
  type Props = {
    message: string;
    variant: "danger" | "info" | "success";
  };

  let { message, variant }: Props = $props();
</script>

<div role="alert" class="p-4 bg-error-bg border border-error-border rounded-2xl">
  {message}
</div>
