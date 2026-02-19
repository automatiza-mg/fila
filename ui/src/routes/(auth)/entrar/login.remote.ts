import { form, getRequestEvent } from "$app/server";
import { env } from "$env/dynamic/public";
import z from "zod/v4";

const schema = z.object({
  cpf: z
    .string()
    .regex(/\d{3}\.\d{3}\.\d{3}\-\d{2}/, "Deve possuir formato 000.000.000-00"),
  senha: z
    .string()
    .min(8, "Deve possuir pelo menos 8 caracteres")
    .max(60, "Deve possuir até 60 caracteres"),
});

export type Token = {
  token: string;
  expira: string;
};

export const login = form(schema, async ({ cpf, senha }) => {
  const { cookies, fetch } = getRequestEvent();

  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/entrar`, {
    method: "POST",
    body: JSON.stringify({
      cpf,
      senha,
    }),
  });

  if (!res.ok) {
    throw new Error("Não foi possível entrar");
  }

  const { expira, token } = (await res.json()) as Token;

  cookies.set("auth_token", token, {
    path: "/",
    expires: new Date(expira),
  });
});
