import { error } from "@sveltejs/kit";

export const load = async ({ url }) => {
  const token = url.searchParams.get("token");
  if (!token) {
    error(401, "Token não informado");
  }
};
