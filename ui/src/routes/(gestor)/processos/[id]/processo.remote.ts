import { command, form } from "$app/server";
import { getClient } from "$lib/server/util";
import { z } from "zod";

const criarPrioridadeSchema = z.object({
  paId: z.string(),
  justificativa: z.string("Deve ser informado"),
});

export const criarPrioridadeForm = form(criarPrioridadeSchema, async (data) => {
  const client = getClient();

  const paId = parseInt(data.paId, 10);
  await client.solicitarPrioridade(paId, {
    justificativa: data.justificativa,
  });
});

const syncPreviewSchema = z.object({
  paId: z.number().int(),
});

export const syncPreviewCmd = command(syncPreviewSchema, async ({ paId }) => {
  const client = getClient();
  await client.syncAposentadoriaPreview(paId);
});
