import { Client } from "$lib/api";
import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
  const authToken = event.cookies.get("auth_token");
  if (authToken) {
    const client = new Client(authToken, event.fetch);
    const usuario = await client.usuarioAtual();
    event.locals.usuario = usuario;
  }

  const response = await resolve(event);
  return response;
};
