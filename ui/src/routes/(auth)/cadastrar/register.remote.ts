import { form, getRequestEvent } from "$app/server";
import { cadastrar } from "$lib/api";
import { invalid } from "@sveltejs/kit";
import { z } from "zod/v4";

const scheme = z.object({
  token: z.string(),
  cpf: z.string(),
  _senha: z
    .string()
    .min(8, "Deve possuir pelo menos 8 caracteres")
    .max(60, "Deve possuir até 60 caracteres"),
  _confirmar_senha: z
    .string()
    .min(8, "Deve possuir pelo menos 8 caracteres")
    .max(60, "Deve possuir até 60 caracteres"),
});

export const registrar = form(
  scheme,
  async ({ token, _senha, _confirmar_senha }, issue) => {
    try {
      await cadastrar({
        token,
        senha: _senha,
        confirmar_senha: _confirmar_senha,
      });
    } catch (err) {
      invalid("Não foi possível concluir o cadastro");
    }
  },
);
