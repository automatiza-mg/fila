import type { Papel, Usuario } from "$lib/api/types.js";

/**
 * Verifica se o usuário possui um dos papeis listados.
 */
export function hasPapel(
  usuario: Usuario | undefined,
  ...allowedPapeis: Papel[]
): boolean {
  if (!usuario || !usuario.papel) {
    return false;
  }

  return allowedPapeis.includes(usuario.papel);
}
