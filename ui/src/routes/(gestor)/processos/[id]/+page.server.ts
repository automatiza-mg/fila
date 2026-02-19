import { error } from "@sveltejs/kit";

export const load = async ({ locals, params }) => {
  const { id } = params;

  const processoId = parseInt(id, 10);
  try {
    const processo = await locals.auth?.client.getAposentadoria(processoId);
    return {
      processo,
    };
  } catch (err) {
    error(404, "Não foi possível buscar o processo");
  }
};
