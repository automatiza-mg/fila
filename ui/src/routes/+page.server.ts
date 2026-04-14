import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  if (!locals.usuario) {
    redirect(302, "/entrar");
  }

  if (locals.usuario.papel === "ANALISTA") {
    redirect(302, "/analise");
  }

  redirect(302, "/processos");
};
