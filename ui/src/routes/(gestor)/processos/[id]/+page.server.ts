import { getClient } from "$lib/server/util";

export const load = async ({ params }) => {
  const { id } = params;
  const processoId = parseInt(id, 10);
  const client = getClient();

  const processo = await client.getAposentadoria(processoId);
  const historico = await client.getHistorico(processoId);

  return {
    processo,
    historico,
  };
};
