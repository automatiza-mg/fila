import { form } from "$app/server";
import { ApiError } from "$lib/api";
import { getClient } from "$lib/server/utils";
import { invalid, redirect } from "@sveltejs/kit";
import z from "zod/v4";

const schema = z.object({
  usuario_id: z.string().min(1),
  orgao: z.enum(["SEPLAG", "SEE"], {
    message: "Deve ser um dos valores: SEPLAG, SEE",
  }),
  sei_unidade_id: z.string().min(1, "Campo obrigatório"),
});

export const criarAnalista = form(
  schema,
  async ({ usuario_id, orgao, sei_unidade_id }, issue) => {
    const client = getClient();
    const id = parseInt(usuario_id, 10);

    try {
      await client.criarAnalista(id, { orgao, sei_unidade_id });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 422 && err.response.errors) {
          const errors = err.response.errors;
          const issues = [];
          for (const [key, value] of Object.entries(errors)) {
            if (key === "orgao") issues.push(issue.orgao(value));
            if (key === "sei_unidade_id")
              issues.push(issue.sei_unidade_id(value));
          }
          invalid(...issues);
        }
        invalid(err.message);
      }
      invalid("Algo deu errado ao cadastrar os dados de analista");
    }

    redirect(303, `/usuarios/${id}`);
  },
);
