<script lang="ts">
  type Props = {
    message: string;
    variant: "danger" | "info" | "success";
  };

  let { message, variant }: Props = $props();
</script>

<div role="alert" class="p-4 bg-red-50 border border-red-300 rounded-2xl">
  {message}
</div>
