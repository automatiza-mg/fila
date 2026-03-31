import { form } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { invalid } from "@sveltejs/kit";
import { z } from "zod";

const leituraInvalidaSchema = z.object({
  processoId: z.coerce.number().int(),
  _motivo: z.string().min(1, "Campo obrigatório"),
});

export const leituraInvalidaForm = form(
  leituraInvalidaSchema,
  async ({ processoId, _motivo }, issue) => {
    const client = getClient();

    try {
      await client.marcarLeituraInvalida(processoId, _motivo);
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
