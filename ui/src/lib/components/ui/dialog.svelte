<script lang="ts">
  import type { Snippet } from "svelte";
  import { Dialog, type WithoutChild } from "bits-ui";
  import { cn } from "$lib/cn";
  import Button from "./button.svelte";

  type Props = Dialog.RootProps & {
    buttonText: string;
    title: Snippet;
    description?: Snippet;
    class?: string;
    contentProps?: WithoutChild<Dialog.ContentProps>;
  };

  let {
    open = $bindable(false),
    children,
    buttonText,
    contentProps,
    title,
    description,
    class: className,
    ...restProps
  }: Props = $props();
</script>

<Dialog.Root bind:open {...restProps}>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <Button {...props}>{buttonText}</Button>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Portal>
    <Dialog.Overlay
      class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
    />
    <Dialog.Content
      class={cn(
        "outline-none bg-white fixed left-[50%] rounded-2xl top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-5 sm:max-w-110 md:w-full data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
        className,
      )}
      {...contentProps}
    >
      <Dialog.Title class="text-lg font-semibold tracking-tight">
        {@render title()}
      </Dialog.Title>
      {#if description}
        <Dialog.Description class="text-muted-foreground text-sm">
          {@render description()}
        </Dialog.Description>
      {/if}
      {@render children?.()}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
