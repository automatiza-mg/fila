import { redirect } from "@sveltejs/kit";

export const load = async ({ parent }) => {
  const { processo } = await parent();

  /**
   * Verifica se o usuário possui um processo ativo, caso contrário envia
   * de volta para a página de análise.
   */
  if (!processo) {
    redirect(302, "/analise");
  }

  return {
    processo,
  };
};
