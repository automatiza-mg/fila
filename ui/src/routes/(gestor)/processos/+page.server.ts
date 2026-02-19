export const load = async ({ locals }) => {
  const processos = await locals.auth?.client.listarAposentadoria();

  return {
    processos,
  };
};
