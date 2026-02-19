// See https://svelte.dev/docs/kit/types#app.d.ts

import type { Usuario } from "$lib/api/types";

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      usuario?: Usuario;
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {};
