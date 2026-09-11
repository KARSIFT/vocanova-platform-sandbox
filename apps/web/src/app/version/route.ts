import { buildIdentity } from "@/lib/build-identity";

export const dynamic = "force-static";

export function GET() {
  return Response.json(buildIdentity, {
    headers: {
      "Cache-Control": "no-store, max-age=0",
    },
  });
}
