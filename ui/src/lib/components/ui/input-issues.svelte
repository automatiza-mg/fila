<script lang="ts">
  type Issue = {
    message: string;
  };

  type Props = {
    issues: Issue[];
  };

  let { issues }: Props = $props();
</script>

{#each issues as issue}
  <p class="text-sm text-error">{issue.message}</p>
{/each}
