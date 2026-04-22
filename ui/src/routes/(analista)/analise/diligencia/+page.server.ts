import type { SolicitacaoDiligencia } from "$lib/api/types";
import { getClient } from "$lib/server/util";
import { redirect } from "@sveltejs/kit";

export const load = async ({ parent }) => {
  const { processo } = await parent();

  /**
   * Verifica se o usuário possui um processo ativo, caso contrário envia
   * de volta para a página de análise.
   */
  if (!processo) {
    redirect(302, "/analise");
  }

  const client = getClient();

  let rascunho: SolicitacaoDiligencia;
  try {
    rascunho = await client.getDiligenciaRascunho(processo.id);
  } catch {
    redirect(302, "/analise");
  }

  return {
    processo,
    rascunho,
  };
};
