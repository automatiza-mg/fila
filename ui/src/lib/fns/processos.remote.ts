import { command, form, query } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { invalid } from "@sveltejs/kit";
import { z } from "zod";

const criarProcessoSchema = z.object({
  numero: z.string().min(1, "Campo obrigatório"),
});

export const criarProcessoForm = form(
  criarProcessoSchema,
  async (data, issue) => {
    const client = getClient();

    try {
      await client.criarProcesso(data);
    } catch (err) {
      if (err instanceof ApiError) {
        invalid(err.message);
      }
      invalid(
        "Não foi possível criar o processo, tente novamente mais tarde",
      );
    }
  },
);

export const recalcularScoresCmd = command("unchecked", async () => {
  const client = getClient();
  await client.recalcularScore();
});

const servidorQuerySchema = z.object({
  cpf: z.string(),
});

export const servidorQuery = query(servidorQuerySchema, async ({ cpf }) => {
  const client = getClient();

  try {
    return await client.getServidor(cpf);
  } catch (err) {
    if (err instanceof ApiError && (err.status === 404 || err.status === 409)) {
      return null;
    }
    throw err;
  }
});

const atualizarPreviewSchema = z.object({
  processoId: z.string(),
});

export const atualizarProcessoPreviewCmd = command(
  atualizarPreviewSchema,
  async ({ processoId }) => {
    const client = getClient();
    await client.atualizarProcessoPreview(processoId);
  },
);
