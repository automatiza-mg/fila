import type { StatusProcessoAposentadoria } from "$lib/api/types";
import { getClient } from "$lib/server/util";

export const load = async ({ url }) => {
  const client = getClient();
  const numero = url.searchParams.get("numero") || undefined;
  const status =
    (url.searchParams.get("status") as StatusProcessoAposentadoria) ||
    undefined;
  const page = parseInt(url.searchParams.get("page") ?? "1", 10) || 1;
  const processos = await client.listarAposentadoria({
    page,
    numero,
    status,
  });

  return {
    processos,
    numero: numero ?? "",
    status: (status ?? "") as StatusProcessoAposentadoria | "",
  };
};
