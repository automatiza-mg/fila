import { getClient } from "$lib/server/util";

export const load = async ({ url }) => {
  const client = getClient();
  const status = url.searchParams.get("status") || undefined;
  const numero = url.searchParams.get("numero") || undefined;
  const page = parseInt(url.searchParams.get("page") ?? "1", 10) || 1;
  const solicitacoes = await client.listarSolicitacoesPrioridade({
    page,
    status,
    numero,
  });

  return {
    solicitacoes,
    status: status ?? "",
    numero: numero ?? "",
  };
};
