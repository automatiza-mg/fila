import { getClient } from "$lib/server/utils.js";
import { error } from "@sveltejs/kit";

export const load = async ({ params }) => {
  const { id } = params;
  const processoId = parseInt(id, 10);
  const client = getClient();

  try {
    const processo = await client.getAposentadoria(processoId);
    return {
      processo,
    };
  } catch (err) {
    error(404, "Não foi possível buscar o processo");
  }
};
