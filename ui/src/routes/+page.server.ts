import { redirect } from "@sveltejs/kit";

export const load = async ({ locals }) => {
  redirect(302, "/entrar");
};
