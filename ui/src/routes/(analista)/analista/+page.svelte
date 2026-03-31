<script lang="ts">
  import NumeroProcesso from "$lib/components/numero-processo.svelte";
  import Button from "$lib/components/ui/button.svelte";
  import { calcularIdade } from "$lib/date";
  import { formatCpf } from "$lib/formatter";
  import { statusText, statusColor } from "$lib/processo";
  import InfoIcon from "phosphor-svelte/lib/InfoIcon";
  import ArrowSquareOutIcon from "phosphor-svelte/lib/ArrowSquareOutIcon";
  import IdentificationCardIcon from "phosphor-svelte/lib/IdentificationCardIcon";
  import CakeIcon from "phosphor-svelte/lib/CakeIcon";
  import CalendarIcon from "phosphor-svelte/lib/CalendarIcon";
  import GavelIcon from "phosphor-svelte/lib/GavelIcon";
  import WheelchairIcon from "phosphor-svelte/lib/WheelchairIcon";
  import FlagIcon from "phosphor-svelte/lib/FlagIcon";
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<svelte:head>
  <title>Analista - Fila Aposentadoria</title>
</svelte:head>

{#if data.processo}
  <div class="space-y-8">
    <div
      class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
    >
      <NumeroProcesso numero={data.processo.numero} />
      <Button
        variant="outline"
        href="/preview/{data.processo.id}"
        target="_blank"
        rel="noopener noreferrer"
      >
        Pré-visualizar Processo
        <ArrowSquareOutIcon />
      </Button>
    </div>

    <div class="space-y-2">
      <div
        class="rounded-xl border border-stone-200 shadow-xs divide-y divide-stone-200 text-sm sm:text-base"
      >
        <div
          class="grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-stone-200"
        >
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <IdentificationCardIcon class="size-3.5 sm:size-4" />
              CPF Requerente
            </p>
            <p class="font-medium mt-0.5">
              {formatCpf(data.processo.cpf_requerente)}
            </p>
          </div>
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <CakeIcon class="size-3.5 sm:size-4" />
              Data de Nascimento
            </p>
            <p class="font-medium mt-0.5">
              {new Date(
                data.processo.data_nascimento_requerente,
              ).toLocaleDateString("pt-BR", { timeZone: "UTC" })}
              <span
                class="text-muted-foreground text-xs sm:text-sm font-normal"
              >
                ({calcularIdade(data.processo.data_nascimento_requerente)} anos)
              </span>
            </p>
          </div>
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <CalendarIcon class="size-3.5 sm:size-4" />
              Data Requerimento
            </p>
            <p class="font-medium mt-0.5">
              {new Date(data.processo.data_requerimento).toLocaleDateString(
                "pt-BR",
                { timeZone: "UTC" },
              )}
            </p>
          </div>
        </div>

        <div
          class="grid grid-cols-2 sm:grid-cols-4 divide-y sm:divide-y-0 sm:divide-x divide-stone-200"
        >
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <InfoIcon class="size-3.5 sm:size-4" />
              Status
            </p>
            <p class="mt-0.5">
              <span
                class="inline-block rounded-md px-2 py-0.5 text-xs sm:text-sm font-medium {statusColor(
                  data.processo.status,
                )}"
              >
                {statusText(data.processo.status)}
              </span>
            </p>
          </div>
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <GavelIcon class="size-3.5 sm:size-4" />
              Judicial
            </p>
            <p class="font-medium mt-0.5">
              {data.processo.judicial ? "Sim" : "Não"}
            </p>
          </div>
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <WheelchairIcon class="size-3.5 sm:size-4" />
              Invalidez
            </p>
            <p class="font-medium mt-0.5">
              {data.processo.invalidez ? "Sim" : "Não"}
            </p>
          </div>
          <div class="px-4 py-3">
            <p
              class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
            >
              <FlagIcon class="size-3.5 sm:size-4" />
              Prioritário
            </p>
            <p class="font-medium mt-0.5">
              {data.processo.prioridade ? "Sim" : "Não"}
            </p>
          </div>
        </div>
      </div>
      <p class="text-xs text-muted-foreground px-1">
        Os dados acima foram extraídos e analisados automaticamente por
        inteligência artificial.
        <span class="font-medium">
          Verifique as informações antes de prosseguir.
        </span>
      </p>
    </div>
  </div>
{:else}
  <div class="flex grow items-center justify-center">
    <p class="text-muted-foreground text-sm">
      Nenhum processo atribuído no momento.
    </p>
  </div>
{/if}
