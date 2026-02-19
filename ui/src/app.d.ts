// See https://svelte.dev/docs/kit/types#app.d.ts

import type { Client } from "$lib/api";
import type { Usuario } from "$lib/api/types";

type AuthContext = {
  client: Client;
  usuario: Usuario;
};

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      auth?: AuthContext;
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {};
