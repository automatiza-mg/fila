import { ApiError } from "$lib/api/client.js";
import type { Analista } from "$lib/api/types.js";
import { getClient } from "$lib/server/util";

export const load = async ({ params, parent }) => {
  const { id } = params;
  const client = getClient();
  const { usuario: usuarioAtual } = await parent();

  const usuarioId = parseInt(id, 10);

  const usuario = await client.getUsuario(usuarioId);

  let analista: Analista | null = null;

  if (usuario.papel === "ANALISTA") {
    try {
      analista = await client.getAnalista(usuario.id);
    } catch (err) {
      const notFound = err instanceof ApiError && err.status === 404;
      if (!notFound) {
        throw err;
      }
    }
  }

  return {
    usuario,
    usuarioAtual,
    analista,
  };
};
