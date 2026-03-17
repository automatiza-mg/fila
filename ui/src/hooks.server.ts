import { Client } from "$lib/api";
import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
  const authToken = event.cookies.get("auth");
  if (authToken) {
    const client = new Client(authToken);
    try {
      const usuario = await client.usuarioAtual();
      event.locals.usuario = usuario;
    } catch (err) {
      event.cookies.delete("auth", {
        path: "/",
      });
    }
  }

  const response = await resolve(event);
  return response;
};
