import { getClient } from "$lib/server/util";

export const load = async ({ url }) => {
  const client = getClient();
  const status = url.searchParams.get("status") || undefined;
  const numero = url.searchParams.get("numero") || undefined;
  const solicitacoes = await client.listarSolicitacoesPrioridade({
    status,
    numero,
  });

  return {
    solicitacoes,
    status: status ?? "",
    numero: numero ?? "",
  };
};
