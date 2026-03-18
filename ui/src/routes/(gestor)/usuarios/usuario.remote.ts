import { form } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { invalid, redirect } from "@sveltejs/kit";
import { z } from "zod";

const createUsuarioSchema = z.object({
  nome: z
    .string()
    .min(1, "Campo obrigatório")
    .max(255, "Deve possuir até 255 caracteres"),
  cpf: z
    .string()
    .regex(
      /^\d{3}\.\d{3}\.\d{3}\-\d{2}$/,
      "Deve possuir formato 000.000.000-00",
    ),
  email: z
    .email("Deve ser um email válido")
    .min(1, "Campo obrigatório")
    .max(255, "Deve possuir até 255 caracteres"),
  papel: z.enum(["ANALISTA", "GESTOR", "SUBSECRETARIO"], {
    message: "Deve ser um dos valores: ANALISTA, GESTOR, SUBSECRETARIO",
  }),
});

export const createUsuarioForm = form(
  createUsuarioSchema,
  async (data, issue) => {
    const client = getClient();

    try {
      await client.criarUsuario(data);
    } catch (err) {
      if (err instanceof ApiError) {
        invalid(err.message);
      }
      invalid("Não foi possível criar o usuário, tente novamente mais tarde");
    }

    redirect(303, "/usuarios");
  },
);
