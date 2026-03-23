import { getClient } from "$lib/server/util";

export const load = async ({ url }) => {
  const client = getClient();
  const numero = url.searchParams.get("numero") || undefined;
  const processos = await client.listarAposentadoria({ numero });

  return {
    processos,
    numero: numero ?? "",
  };
};
