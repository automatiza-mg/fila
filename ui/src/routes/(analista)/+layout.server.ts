import { ApiError } from "$lib/api/client.js";
import type { ProcessoAposentadoria } from "$lib/api/types.js";
import { hasPapel } from "$lib/papel.js";
import { getClient } from "$lib/server/util";
import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  if (!locals.usuario || !hasPapel(locals.usuario, "ANALISTA")) {
    redirect(302, "/");
  }

  const client = getClient();

  let processo: ProcessoAposentadoria | null = null;
  try {
    processo = await client.meuProcessoAtribuido();
  } catch (err) {
    const notFound = err instanceof ApiError && err.status === 404;
    if (!notFound) {
      throw err;
    }
  }

  return {
    usuario: locals.usuario,
    processo,
  };
};
