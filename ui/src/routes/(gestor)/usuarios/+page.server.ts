export const load = async ({ locals }) => {
  const usuarios = await locals.auth?.client.listarUsuarios();

  return {
    usuarios,
  };
};
