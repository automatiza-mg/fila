import { error } from "@sveltejs/kit";
import { tokenInfo, ApiError } from "$lib/api/client";

export const load = async ({ url }) => {
  const token = url.searchParams.get("token");
  if (!token) {
    error(401, "Token não informado");
  }

  try {
    const usuario = await tokenInfo(token, "reset-senha");
    return { token, usuario };
  } catch (e) {
    if (e instanceof ApiError) {
      error(e.status, e.message);
    }
    throw e;
  }
};
