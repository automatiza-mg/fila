import { form } from "$app/server";
import { ApiError } from "$lib/api";
import { getClient } from "$lib/server/utils";
import { invalid, redirect } from "@sveltejs/kit";
import z from "zod/v4";

const schema = z.object({
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
    .string()
    .min(1, "Campo obrigatório")
    .max(255, "Deve possuir até 255 caracteres")
    .email("Deve ser um email válido"),
  papel: z.enum(["ANALISTA", "GESTOR", "SUBSECRETARIO"], {
    message: "Deve ser um dos valores: ANALISTA, GESTOR, SUBSECRETARIO",
  }),
});

export const criar = form(schema, async ({ nome, cpf, email, papel }, issue) => {
  const client = getClient();

  try {
    await client.criarUsuario({ nome, cpf, email, papel });
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 422 && err.response.errors) {
        const errors = err.response.errors;
        const issues = [];
        for (const [key, value] of Object.entries(errors)) {
          if (key === "nome") issues.push(issue.nome(value));
          if (key === "cpf") issues.push(issue.cpf(value));
          if (key === "email") issues.push(issue.email(value));
          if (key === "papel") issues.push(issue.papel(value));
        }
        invalid(...issues);
      }
      invalid(err.message);
    }
    invalid("Algo deu errado ao criar o usuário");
  }

  redirect(303, "/usuarios");
});
