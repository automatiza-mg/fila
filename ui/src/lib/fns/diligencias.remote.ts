import { command } from "$app/server";
import { ApiError } from "$lib/api/client";
import type { SolicitacaoDiligencia } from "$lib/api/types";
import { getClient } from "$lib/server/util";
import { error } from "@sveltejs/kit";
import { z } from "zod";

const itemSchema = z.object({
  tipo: z.string().min(1, "Campo obrigatório"),
  subcategorias: z.array(z.string()),
  detalhe: z.string(),
});

const salvarSchema = z.object({
  paId: z.number(),
  itens: z.array(itemSchema),
});

export const salvarDiligencia = command(
  salvarSchema,
  async ({ paId, itens }): Promise<SolicitacaoDiligencia> => {
    const client = getClient();
    try {
      return await client.salvarDiligenciaRascunho(paId, { itens });
    } catch (err) {
      if (err instanceof ApiError) {
        error(err.status, err.message);
      }
      error(500, "Não foi possível salvar as diligências");
    }
  },
);

const enviarSchema = z.object({ paId: z.number() });

export const enviarDiligencia = command(
  enviarSchema,
  async ({ paId }): Promise<SolicitacaoDiligencia> => {
    const client = getClient();
    try {
      return await client.enviarDiligenciaRascunho(paId);
    } catch (err) {
      if (err instanceof ApiError) {
        error(err.status, err.message);
      }
      error(500, "Não foi possível enviar a diligência");
    }
  },
);

const descartarSchema = z.object({ paId: z.number() });

export const descartarDiligencia = command(
  descartarSchema,
  async ({ paId }): Promise<void> => {
    const client = getClient();
    try {
      await client.descartarDiligenciaRascunho(paId);
    } catch (err) {
      if (err instanceof ApiError) {
        error(err.status, err.message);
      }
      error(500, "Não foi possível descartar o rascunho");
    }
  },
);
