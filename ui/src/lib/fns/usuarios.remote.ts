import { command, form, getRequestEvent, query } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { invalid } from "@sveltejs/kit";
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
  },
);

const analistaQuerySchema = z.object({
  usuarioId: z.number().int(),
});

export const analistaQuery = query(
  analistaQuerySchema,
  async ({ usuarioId }) => {
    const client = getClient();

    try {
      return await client.getAnalista(usuarioId);
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        return null;
      }
      throw err;
    }
  },
);

const deleteUsuarioSchema = z.object({
  usuarioId: z.number().int(),
});

export const deleteUsuarioCmd = command(
  deleteUsuarioSchema,
  async ({ usuarioId }) => {
    const client = getClient();
    const { locals } = getRequestEvent();

    if (usuarioId === locals.usuario?.id) {
      throw new Error("Não é possível excluir sua própria conta.");
    }

    await client.deletarUsuario(usuarioId);
  },
);

const enviarCadastroSchema = z.object({
  usuarioId: z.number().int(),
});

export const enviarCadastroCmd = command(
  enviarCadastroSchema,
  async ({ usuarioId }) => {
    const client = getClient();
    await client.enviarCadastro(usuarioId);
  },
);

export const unidadesQuery = query(async () => {
  const client = getClient();
  return await client.listarUnidadesSei();
});

const criarAnalistaSchema = z.object({
  usuarioId: z.coerce.number().int(),
  orgao: z.string().min(1, "Campo obrigatório"),
  sei_unidade_id: z.string().min(1, "Campo obrigatório"),
});

export const criarAnalistaForm = form(
  criarAnalistaSchema,
  async ({ usuarioId, orgao, sei_unidade_id }, issue) => {
    const client = getClient();

    try {
      await client.criarAnalista(usuarioId, { sei_unidade_id, orgao });
    } catch (err) {
      if (err instanceof ApiError) {
        invalid(err.message);
      }
      invalid(
        "Não foi possível cadastrar o analista, tente novamente mais tarde",
      );
    }
  },
);

const afastarAnalistaSchema = z.object({
  usuarioId: z.number().int(),
});

export const afastarAnalistaCmd = command(
  afastarAnalistaSchema,
  async ({ usuarioId }) => {
    const client = getClient();
    await client.afastarAnalista(usuarioId);
  },
);

const retornarAnalistaSchema = z.object({
  usuarioId: z.number().int(),
});

export const retornarAnalistaCmd = command(
  retornarAnalistaSchema,
  async ({ usuarioId }) => {
    const client = getClient();
    await client.retornarAnalista(usuarioId);
  },
);
