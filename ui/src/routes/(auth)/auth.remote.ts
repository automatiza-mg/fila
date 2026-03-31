import { form, getRequestEvent } from "$app/server";
import { env } from "$env/dynamic/public";
import {
  cadastrar,
  entrar,
  recuperarSenha,
  redefinirSenha,
  ApiError,
} from "$lib/api/client";
import { error, invalid, redirect } from "@sveltejs/kit";
import { z } from "zod/v4";

const entrarSchema = z.object({
  cpf: z.string(),
  _senha: z.string(),
});

export const entrarForm = form(entrarSchema, async ({ cpf, _senha }, issue) => {
  const { cookies } = getRequestEvent();

  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/entrar`, {
    method: "POST",
    body: JSON.stringify({
      cpf: cpf,
      senha: _senha,
    }),
  });

  if (res.ok) {
    const { token, expira } = await res.json();
    cookies.set("auth", token, {
      path: "/",
      expires: new Date(expira),
      httpOnly: true,
    });

    redirect(303, "/painel");
  }

  invalid("Não foi possível autenticar, confira suas credenciais.");
});

const recuperarSenhaSchema = z.object({
  cpf: z.string(),
});

export const recuperarSenhaForm = form(
  recuperarSenhaSchema,
  async ({ cpf }) => {
    await recuperarSenha({ cpf });
  },
);

const redefinirSenhaSchema = z.object({
  token: z.string(),
  _senha: z.string(),
  _confirmar_senha: z.string(),
});

export const redefinirSenhaForm = form(
  redefinirSenhaSchema,
  async ({ token, _senha, _confirmar_senha }, issue) => {
    try {
      await redefinirSenha({
        token,
        senha: _senha,
        confirmar_senha: _confirmar_senha,
      });
    } catch (e) {
      if (e instanceof ApiError && e.response?.errors) {
        for (const [, message] of Object.entries(e.response.errors)) {
          issue(message);
        }
        return;
      }
      if (e instanceof ApiError) {
        issue(e.message);
        return;
      }
      throw e;
    }

    redirect(303, "/entrar");
  },
);

const cadastrarSchema = z.object({
  token: z.string(),
  cpf: z.string(),
  _senha: z.string(),
  _confirmar_senha: z.string(),
});

export const cadastrarForm = form(
  cadastrarSchema,
  async ({ token, cpf, _senha, _confirmar_senha }, issue) => {
    try {
      await cadastrar({
        token,
        senha: _senha,
        confirmar_senha: _confirmar_senha,
      });
    } catch (e) {
      if (e instanceof ApiError && e.response?.errors) {
        for (const [, message] of Object.entries(e.response.errors)) {
          issue(message);
        }
        return;
      }
      if (e instanceof ApiError) {
        issue(e.message);
        return;
      }
      throw e;
    }

    // Autenticar automaticamente após o cadastro.
    try {
      const { cookies } = getRequestEvent();
      const { token: authToken, expira } = await entrar({
        cpf,
        senha: _senha,
      });

      console.log(token, expira);
      cookies.set("auth", authToken, {
        path: "/",
        expires: new Date(expira),
        httpOnly: true,
      });
    } catch (e) {
      console.log(e);
      // Cadastro concluído, mas login falhou — redirecionar para entrar.
      redirect(303, "/entrar");
    }

    redirect(303, "/");
  },
);

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
