import { form, getRequestEvent } from "$app/server";
import { error, redirect } from "@sveltejs/kit";

export const sairForm = form("unchecked", async () => {
  const { locals, cookies } = getRequestEvent();
  if (!locals.usuario) {
    error(401, "Usuário não autenticado");
  }
  cookies.delete("auth", {
    path: "/",
  });
  redirect(303, "/entrar");
});
