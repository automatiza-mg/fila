<script lang="ts">
  import type { Snippet } from "svelte";
  import Label from "./label.svelte";
  import InputIssues from "./input-issues.svelte";

  type Props = {
    label: string;
    id?: string;
    issues?: { message: string }[];
    children: Snippet;
  };

  let { label, id, issues = [], children }: Props = $props();
</script>

<div class="grid gap-1">
  <Label for={id}>{label}</Label>
  {@render children()}
  <InputIssues {issues} />
</div>
