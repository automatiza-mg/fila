import { command, form } from "$app/server";
import { ApiError } from "$lib/api/client";
import { getClient } from "$lib/server/util";
import { invalid } from "@sveltejs/kit";
import { z } from "zod";

const criarProcessoSchema = z.object({
  numero: z.string().min(1, "Campo obrigatório"),
});

export const recalcularScoresCmd = command("unchecked", async () => {
  const client = getClient();
  await client.recalcularScore();
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
