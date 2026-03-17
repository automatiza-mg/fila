import { getClient } from "$lib/server/util";

export const load = async () => {
  const client = getClient();
  const processos = await client.listarAposentadoria();

  return {
    processos,
  };
};
