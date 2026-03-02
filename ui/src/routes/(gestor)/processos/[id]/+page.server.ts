import { getClient } from "$lib/server/utils.js";
import type { Documento, Processo, ProcessoHistorico } from "$lib/api/types.js";
import { error } from "@sveltejs/kit";

export const load = async ({ params }) => {
  const { id } = params;
  const processoId = parseInt(id, 10);
  const client = getClient();

  try {
    const processo = await client.getAposentadoria(processoId);

    const [historico, processoBase] = await Promise.all([
      client.getProcessoAposentadoriaHistorico(processoId).catch(() => [] as ProcessoHistorico[]),
      client.getProcesso(processo.processo_id).catch(() => null as Processo | null),
    ]);

    let documentos: Documento[] = [];
    if (processoBase) {
      documentos = await client.getProcessoDocumentos(processoBase.id).catch(() => [] as Documento[]);
    }

    return {
      processo,
      historico,
      processoBase,
      documentos,
    };
  } catch {
    error(404, "Não foi possível buscar o processo");
  }
};
