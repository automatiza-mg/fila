import { getRequestEvent } from "$app/server";
import { Client } from "$lib/api";

export function getClient() {
  const event = getRequestEvent();
  const authToken = event.cookies.get("auth_token");
  if (!authToken) {
    throw new Error("Not authenticated");
  }

  return new Client(authToken, event.fetch);
}
