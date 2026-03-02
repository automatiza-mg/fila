import { form } from "$app/server";
import { ApiError } from "$lib/api";
import { getClient } from "$lib/server/utils";
import { invalid, redirect } from "@sveltejs/kit";
import z from "zod/v4";

const schema = z.object({
  usuario_id: z.string().min(1),
});

export const excluir = form(schema, async ({ usuario_id }, _issue) => {
  const client = getClient();
  const id = parseInt(usuario_id, 10);

  try {
    await client.deleteUsuario(id);
  } catch (err) {
    if (err instanceof ApiError) {
      invalid(err.message);
    }
    invalid("Algo deu errado ao excluir o usuário");
  }

  redirect(303, "/usuarios");
});
