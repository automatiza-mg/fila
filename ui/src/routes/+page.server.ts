import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  if (!locals.usuario) {
    redirect(302, "/entrar");
  }

  if (locals.usuario.papel === "ANALISTA") {
    redirect(302, "/analista");
  }

  redirect(302, "/painel");
};
