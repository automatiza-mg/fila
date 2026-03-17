import { getClient } from "$lib/server/util";

export const load = async () => {
  const client = getClient();
  const usuarios = await client.listarUsuarios();

  return {
    usuarios,
  };
};
