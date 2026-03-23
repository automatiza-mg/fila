<script lang="ts">
  import type { Snippet } from "svelte";
  import { AlertDialog, type WithoutChild } from "bits-ui";
  import { cn } from "$lib/cn";

  type Props = AlertDialog.RootProps & {
    buttonText: string;
    title: Snippet;
    description: Snippet;
    variant?: "base" | "destructive";
    contentProps?: WithoutChild<AlertDialog.ContentProps>;
    onConfirmed?: () => void;
    onCancelled?: () => void;
  };

  let {
    open = $bindable(false),
    variant = "base",
    children,
    buttonText,
    contentProps,
    title,
    description,
    onConfirmed,
    onCancelled,
    ...restProps
  }: Props = $props();
</script>

<AlertDialog.Root bind:open {...restProps}>
  <AlertDialog.Trigger>
    {buttonText}
  </AlertDialog.Trigger>
  <AlertDialog.Portal>
    <AlertDialog.Overlay
      class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
    />
    <AlertDialog.Content
      interactOutsideBehavior="close"
      class="outline-none bg-white fixed left-[50%] rounded-2xl top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-5 sm:max-w-105 md:w-full data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95"
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
        <AlertDialog.Cancel
          onclick={onCancelled}
          class="text-sm px-4 py-2 hover:bg-stone-100 font-medium border rounded-md border-stone-300 outline-none focus-visible:ring-3 focus-visible:ring-secondary/50 focus-visible:border-secondary"
        >
          Cancelar
        </AlertDialog.Cancel>
        <AlertDialog.Action
          onclick={onConfirmed}
          class={cn(
            "text-sm px-4 py-2 font-medium text-white rounded-md border border-transparent outline-none focus-visible:ring-3",
            variant === "destructive"
              ? "bg-red-700 focus-visible:ring-red-400/50 hover:bg-red-700/90"
              : "bg-primary focus-visible:ring-secondary/50 hover:bg-primary/90",
          )}
        >
          Continuar
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>
