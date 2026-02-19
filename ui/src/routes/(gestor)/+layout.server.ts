import { hasPapel } from "$lib/auth.js";
import { redirect } from "@sveltejs/kit";
import type { Papel } from "$lib/api/types.js";

const allowedPapeis: Papel[] = ["ADMIN", "GESTOR", "SUBSECRETARIO"];

export const load = async ({ locals, cookies }) => {
  const usuario = locals.usuario;
  if (!usuario) {
    redirect(303, "/entrar");
  }

  // TODO: Deslogar usuário ou mostrar página 403.
  if (!hasPapel(usuario, ...allowedPapeis)) {
    cookies.delete("auth_token", {
      path: "/",
    });
    redirect(303, "/entrar");
  }

  return {
    usuario,
  };
};
