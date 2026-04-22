import { getClient } from "$lib/server/util";

export const load = async ({ url }) => {
  const client = getClient();
  const numero = url.searchParams.get("numero") || undefined;
  const page = parseInt(url.searchParams.get("page") ?? "1", 10) || 1;
  const processos = await client.meuHistorico({ page, numero });

  return {
    processos,
    numero: numero ?? "",
  };
};
