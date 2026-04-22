import { command, form } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { error, invalid } from "@sveltejs/kit";
import { z } from "zod";

const leituraInvalidaSchema = z.object({
  processoId: z.string(),
  _motivo: z.string().min(1, "Campo obrigatório"),
});

export const leituraInvalidaForm = form(
  leituraInvalidaSchema,
  async ({ processoId, _motivo }, issue) => {
    const client = getClient();
    const paId = parseInt(processoId, 10);
    try {
      await client.marcarLeituraInvalida(paId, _motivo);
    } catch (err) {
      if (err instanceof ApiError) {
        invalid(err.message);
      }
      invalid(
        "Não foi possível marcar o processo como leitura inválida, tente novamente mais tarde",
      );
    }
  },
);

const registrarPublicacaoSchema = z.object({ paId: z.number() });

export const registrarPublicacao = command(
  registrarPublicacaoSchema,
  async ({ paId }): Promise<void> => {
    const client = getClient();
    try {
      await client.registrarPublicacao(paId);
    } catch (err) {
      if (err instanceof ApiError) {
        error(err.status, err.message);
      }
      error(500, "Não foi possível registrar a publicação");
    }
  },
);
