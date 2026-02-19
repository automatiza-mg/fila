import { getClient } from "$lib/server/utils.js";

export const load = async () => {
  const client = getClient();
  const processos = await client.listarAposentadoria();

  return {
    processos,
  };
};
