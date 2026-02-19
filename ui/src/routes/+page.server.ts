import { hasPapel } from "$lib/auth.js";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  const usuario = locals.usuario;
  if (!usuario) {
    redirect(302, "/entrar");
  }

  if (usuario.papel === "ANALISTA") {
    redirect(302, "/analise");
  }

  if (hasPapel(usuario, "ADMIN", "GESTOR", "SUBSECRETARIO")) {
    redirect(302, "/processos");
  }

  error(403, "Você não tem permissão para acessar essa aplicação.");
};
