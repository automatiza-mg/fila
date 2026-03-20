import type { Papel, Usuario } from "./api/types";

/**
 * Reporta se o usuário possui um dos papeis informados. Sempre retorna
 * `true` para usuários com papel `ADMIN`.
 */
export function hasPapel(usuario: Usuario, ...papeis: Papel[]): boolean {
  if (usuario.papel === "ADMIN") return true;
  return !!usuario.papel && papeis.includes(usuario.papel);
}
