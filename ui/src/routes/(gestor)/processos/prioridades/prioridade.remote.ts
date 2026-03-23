import { command } from "$app/server";
import { getClient } from "$lib/server/util";
import { z } from "zod";

const prioridadeSchema = z.object({
  id: z.number().int(),
});

export const aprovarPrioridadeCmd = command(
  prioridadeSchema,
  async ({ id }) => {
    const client = getClient();
    await client.aprovarSolicitacaoPrioridade(id);
  },
);

export const negarPrioridadeCmd = command(
  prioridadeSchema,
  async ({ id }) => {
    const client = getClient();
    await client.negarSolicitacaoPrioridade(id);
  },
);
