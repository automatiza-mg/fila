import { ApiError, tokenInfo } from "$lib/api";
import { error } from "@sveltejs/kit";

export const load = async ({ url }) => {
  const token = url.searchParams.get("token");
  if (!token) {
    error(401, "Token não informado");
  }

  try {
    const usuario = await tokenInfo(token, "setup");
    return {
      usuario,
      token,
    };
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        error(401, "Token informado é inválido ou expirou");
      }
    }
    error(
      500,
      "Não foi possível validar o token informado, tente novamente mais tarde",
    );
  }
};
