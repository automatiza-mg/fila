<script lang="ts">
  import type { ProcessoAposentadoria } from "$lib/api/types";
  import ServidorPopover from "$lib/components/servidor-popover.svelte";
  import { calcularIdade } from "$lib/date";
  import { statusText, statusColor } from "$lib/processo";
  import InfoIcon from "phosphor-svelte/lib/InfoIcon";
  import IdentificationCardIcon from "phosphor-svelte/lib/IdentificationCardIcon";
  import CakeIcon from "phosphor-svelte/lib/CakeIcon";
  import CalendarIcon from "phosphor-svelte/lib/CalendarIcon";
  import GavelIcon from "phosphor-svelte/lib/GavelIcon";
  import WheelchairIcon from "phosphor-svelte/lib/WheelchairIcon";
  import FlagIcon from "phosphor-svelte/lib/FlagIcon";

  type Props = {
    processo: ProcessoAposentadoria;
  };

  let { processo }: Props = $props();
</script>

<div class="space-y-2">
  <div
    class="rounded-xl border border-border shadow-xs divide-y divide-border text-sm sm:text-base"
  >
    <div
      class="grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-border"
    >
      <div class="px-4 py-3">
        <p
          class="text-muted-foreground text-xs sm:text-sm flex items-center gap-1"
        >
          <IdentificationCardIcon class="size-3.5 sm:size-4" />
          Dados Requerente
        </p>
        <p class="font-medium mt-0.5">
          <ServidorPopover cpf={processo.cpf_requerente} />
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
          {new Date(processo.data_nascimento_requerente).toLocaleDateString(
            "pt-BR",
            { timeZone: "UTC" },
          )}
          <span class="text-muted-foreground text-xs sm:text-sm font-normal">
            ({calcularIdade(processo.data_nascimento_requerente)} anos)
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
          {new Date(processo.data_requerimento).toLocaleDateString("pt-BR", {
            timeZone: "UTC",
          })}
        </p>
      </div>
    </div>

    <div
      class="grid grid-cols-2 sm:grid-cols-4 divide-y sm:divide-y-0 sm:divide-x divide-border"
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
              processo.status,
            )}"
          >
            {statusText(processo.status)}
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
          {processo.judicial ? "Sim" : "Não"}
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
          {processo.invalidez ? "Sim" : "Não"}
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
          {processo.prioridade ? "Sim" : "Não"}
        </p>
      </div>
    </div>
  </div>
  <p class="text-xs text-muted-foreground px-1">
    Os dados acima foram extraídos e analisados automaticamente por inteligência
    artificial.
    <span class="font-medium">
      Verifique as informações antes de prosseguir.
    </span>
  </p>
</div>
