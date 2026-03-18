import { ApiError } from "$lib/api/client.js";
import type { Analista, Unidade } from "$lib/api/types.js";
import { getClient } from "$lib/server/util";

export const load = async ({ params }) => {
  const { id } = params;
  const client = getClient();

  const usuarioId = parseInt(id, 10);

  const usuario = await client.getUsuario(usuarioId);

  let analista: Analista | null = null;
  let unidades: Unidade[] | null = null;

  if (usuario.papel === "ANALISTA") {
    unidades = await client.listarUnidadesSei();
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
    analista,
    unidades,
  };
};
