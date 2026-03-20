import { getClient } from "$lib/server/util";

export const load = async () => {
  const client = getClient();
  const solicitacoes = await client.listarSolicitacoesPrioridade();

  return {
    solicitacoes,
  };
};
