import { hasPapel } from "$lib/papel.js";
import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  // Apenas gestores e subsecretários podem acessar essas rotas.
  if (!locals.usuario || !hasPapel(locals.usuario, "GESTOR", "SUBSECRETARIO")) {
    redirect(302, "/entrar");
  }

  return {
    usuario: locals.usuario,
  };
};
