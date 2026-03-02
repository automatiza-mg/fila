import { getClient } from "$lib/server/utils.js";
import type { Analista } from "$lib/api/types.js";
import { error } from "@sveltejs/kit";

export const load = async ({ params }) => {
  const { id } = params;
  const usuarioId = parseInt(id, 10);
  const client = getClient();

  try {
    const usuario = await client.getUsuario(usuarioId);

    let analista: Analista | null = null;
    if (usuario.papel === "ANALISTA") {
      try {
        analista = await client.getAnalista(usuarioId);
      } catch {
        // Analista data may not exist yet
      }
    }

    return {
      usuario,
      analista,
    };
  } catch {
    error(404, "Não foi possível buscar o usuário");
  }
};
