<script lang="ts">
  import type { Snippet } from "svelte";
  import { AlertDialog, type WithoutChild } from "bits-ui";
  import Button from "./button.svelte";

  type Props = AlertDialog.RootProps & {
    buttonText: string;
    buttonIcon?: Snippet;
    title: Snippet;
    description: Snippet;
    variant?: "default" | "destructive" | "outline";
    disabled?: boolean;
    contentProps?: WithoutChild<AlertDialog.ContentProps>;
    onConfirmed?: () => void;
    onCancelled?: () => void;
  };

  let {
    open = $bindable(false),
    variant = "default",
    disabled = false,
    children,
    buttonText,
    buttonIcon,
    contentProps,
    title,
    description,
    onConfirmed,
    onCancelled,
    ...restProps
  }: Props = $props();
</script>

<AlertDialog.Root bind:open {...restProps}>
  <AlertDialog.Trigger {disabled}>
    {#snippet child({ props })}
      <Button {variant} {disabled} {...props}>
        {#if buttonIcon}{@render buttonIcon()}{/if}
        {buttonText}
      </Button>
    {/snippet}
  </AlertDialog.Trigger>
  <AlertDialog.Portal>
    <AlertDialog.Overlay
      class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-overlay"
    />
    <AlertDialog.Content
      interactOutsideBehavior="close"
      class="outline-none bg-surface fixed left-[50%] rounded-2xl top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-5 sm:max-w-105 md:w-full data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95"
      {...contentProps}
    >
      <AlertDialog.Title class="text-lg font-semibold tracking-tight">
        {@render title()}
      </AlertDialog.Title>
      <AlertDialog.Description class="text-muted-foreground text-sm">
        {@render description()}
      </AlertDialog.Description>
      {@render children?.()}

      <div class="flex justify-end gap-2 items-center mt-4">
        <AlertDialog.Cancel onclick={onCancelled}>
          {#snippet child({ props })}
            <Button variant="outline" {...props}>Cancelar</Button>
          {/snippet}
        </AlertDialog.Cancel>
        <AlertDialog.Action onclick={onConfirmed}>
          {#snippet child({ props })}
            <Button variant={variant === "destructive" ? "destructive" : "default"} {...props}>Continuar</Button>
          {/snippet}
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>
