import { form, getRequestEvent } from "$app/server";
import { ApiError, authenticate } from "$lib/api";
import { invalid, redirect } from "@sveltejs/kit";
import z from "zod/v4";

const schema = z.object({
  cpf: z
    .string()
    .regex(/\d{3}\.\d{3}\.\d{3}\-\d{2}/, "Deve possuir formato 000.000.000-00"),
  _senha: z
    .string()
    .min(8, "Deve possuir pelo menos 8 caracteres")
    .max(60, "Deve possuir até 60 caracteres"),
});

export type Token = {
  token: string;
  expira: string;
};

export const login = form(schema, async ({ cpf, _senha }, issue) => {
  const { cookies, fetch } = getRequestEvent();

  try {
    const { expira, token } = await authenticate({ cpf, senha: _senha });
    cookies.set("auth_token", token, {
      path: "/",
      expires: new Date(expira),
    });
  } catch (err) {
    if (err instanceof ApiError) {
      switch (err.status) {
        case 422:
          const errors = err.response.errors!;
          const issues = [];
          for (const [key, value] of Object.entries(errors)) {
            if (key === "cpf") {
              issues.push(issue.cpf(value));
            }
            if (key === "senha") {
              issues.push(issue._senha(value));
            }
          }
          invalid(...issues);
        default:
          invalid(err.message);
      }
    } else {
      invalid("Algo deu errado ao autenticar");
    }
  }

  redirect(303, "/");
});
