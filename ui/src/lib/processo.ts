import type { StatusProcessoAposentadoria } from "./api/types";

export function statusText(status: StatusProcessoAposentadoria): string {
  switch (status) {
    case "ANALISE_PENDENTE":
      return "Análise Pendente";
    case "CONCLUIDO":
      return "Concluído";
    case "EM_ANALISE":
      return "Em Análise";
    case "EM_DILIGENCIA":
      return "Em Diligência";
    case "LEITURA_INVALIDA":
      return "Leitura Inválida";
    case "RETORNO_DILIGENCIA":
      return "Retorno Diligência";
    default:
      return "Desconhecido";
  }
}

export function statusColor(status: StatusProcessoAposentadoria): string {
  switch (status) {
    case "ANALISE_PENDENTE":
      return "bg-status-pending/50";
    case "CONCLUIDO":
      return "bg-status-completed/50";
    case "EM_ANALISE":
      return "bg-status-analyzing/50";
    case "EM_DILIGENCIA":
      return "bg-status-diligence/50";
    case "LEITURA_INVALIDA":
      return "bg-status-invalid/50";
    case "RETORNO_DILIGENCIA":
      return "bg-status-return/50";
  }
}
