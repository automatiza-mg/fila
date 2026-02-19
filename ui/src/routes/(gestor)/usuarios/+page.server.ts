import { getClient } from "$lib/server/utils.js";

export const load = async () => {
  const client = getClient();
  const usuarios = await client.listarUsuarios();

  return {
    usuarios,
  };
};
