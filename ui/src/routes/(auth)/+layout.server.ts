import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  // Se usuário está autenticado, redireciona para página correta.
  if (locals.usuario) {
    redirect(302, "/");
  }
};
