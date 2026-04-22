<script lang="ts">
  import type { Snippet } from "svelte";
  import type {
    HTMLAnchorAttributes,
    HTMLButtonAttributes,
  } from "svelte/elements";
  import { cn } from "$lib/cn";

  type BaseProps = {
    children: Snippet;
    variant?: "default" | "destructive" | "outline" | "link";
    class?: string;
  };

  type ButtonProps = BaseProps & HTMLButtonAttributes & { href?: never };
  type AnchorProps = BaseProps & HTMLAnchorAttributes & { href: string };
  type Props = ButtonProps | AnchorProps;

  let {
    children,
    variant = "default",
    class: className,
    ...restProps
  }: Props = $props();

  const base =
    "px-4 py-2 text-sm font-medium rounded-lg inline-flex justify-center items-center gap-2 [&>svg]:size-4.5 [&>svg]:-ml-1 [&>svg]:pointer-events-none [&>svg]:shrink-0 outline-none focus-visible:ring-3 focus-visible:ring-secondary/50 disabled:bg-muted disabled:text-muted-foreground disabled:border-border disabled:pointer-events-none";
  const variants = {
    default:
      "bg-primary text-primary-foreground hover:bg-primary/90 border border-transparent",
    destructive:
      "bg-destructive text-destructive-foreground hover:bg-destructive/90 focus-visible:ring-destructive/50",
    outline:
      "border border-border-strong hover:bg-surface-alt focus-visible:border-secondary",
    link: "text-muted-foreground underline hover:text-foreground p-0 rounded-none",
  };
</script>

{#if "href" in restProps && restProps.href}
  <a
    class={cn(base, variants[variant], className)}
    {...restProps as HTMLAnchorAttributes}
  >
    {@render children()}
  </a>
{:else}
  <button
    class={cn(base, variants[variant], className)}
    {...restProps as HTMLButtonAttributes}
  >
    {@render children()}
  </button>
{/if}
