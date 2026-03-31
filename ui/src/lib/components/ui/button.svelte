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
    "px-4 py-2 text-sm font-medium rounded-lg inline-flex items-center gap-2 [&>svg]:size-5 [&>svg]:pointer-events-none [&>svg]:shrink-0";
  const variants = {
    default:
      "bg-primary text-white hover:bg-primary/90 border border-transparent",
    destructive: "bg-red-700 text-white hover:bg-red-700/90",
    outline: "border border-stone-300 hover:bg-stone-100",
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
