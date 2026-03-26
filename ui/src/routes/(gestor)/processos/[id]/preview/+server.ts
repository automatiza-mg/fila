import { getClient } from "$lib/server/util";
import { error } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ params }) => {
  const client = getClient();
  const paId = parseInt(params.id, 10);

  try {
    const blob = await client.getAposentadoriaPreview(paId);
    return new Response(blob, {
      headers: {
        "Content-Type": blob.type || "application/pdf",
        "Content-Disposition": `inline; filename="processo-${params.id}.pdf"`,
      },
    });
  } catch {
    error(404, "Preview não disponível");
  }
};
