import type { StatusProcessoAposentadoria } from "./api/types";

export function statusProcesso(status: StatusProcessoAposentadoria): string {
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
