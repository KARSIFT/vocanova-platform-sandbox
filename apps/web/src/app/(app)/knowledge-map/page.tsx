import { redirect } from "next/navigation";

import { validBand } from "./models";

/**
 * The word map is part of Journey. Earlier `/knowledge-map` links keep working:
 * they open the same range there (the browser keeps any `#range-words`).
 */
export default async function KnowledgeMapPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const band = validBand((await searchParams).band);
  redirect(`/discover?band=${band}`);
}
