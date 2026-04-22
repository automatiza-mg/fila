import { ApiError } from "$lib/api/client";
import type { Analista } from "$lib/api/types";
import { getClient } from "$lib/server/util";
import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  if (!locals.usuario) {
    redirect(302, "/entrar");
  }

  let analista: Analista | null = null;
  if (locals.usuario.papel === "ANALISTA") {
    const client = getClient();
    try {
      analista = await client.analistaAtual();
    } catch (err) {
      if (!(err instanceof ApiError) || err.status !== 404) {
        throw err;
      }
    }
  }

  return {
    usuario: locals.usuario,
    analista,
  };
};
