// Synthetic requester-owned preference storage; database tests prove durability.
export async function handleDiscoveryPosition({ req, res, url, cookies, state, jsonResponse, readJsonBody, checkCsrf, logLine }) {
  if (url.pathname !== "/api/v1/discovery-position") return false;
  res.setHeader("Cache-Control", "no-store");
  const send = (status, body) => { jsonResponse(res, status, body); return true; };
  if (req.method === "GET") return send(200, { position: state.discoveryPosition ?? null });
  if (req.method !== "PUT") return send(405, { detail: "Method not allowed" });
  if (!checkCsrf(req, cookies, res, logLine)) return true;
  let body;
  try { body = await readJsonBody(req); } catch { return send(400, { detail: "Invalid position" }); }
  if (body.catalogVersion !== "synthetic-browser-v1" || !Number.isInteger(body.rank) || body.rank < 1 || body.rank > 10000 || body.band !== Math.floor((body.rank - 1) / 100) + 1 || typeof body.active !== "boolean" || !Number.isSafeInteger(body.expectedRevision) || body.expectedRevision < 0) return send(400, { detail: "Invalid position" });
  if (body.expectedRevision !== (state.discoveryPosition?.revision ?? 0)) return send(409, { detail: "Position changed elsewhere" });
  state.discoveryPosition = { catalogVersion: body.catalogVersion, band: body.band, rank: body.rank, active: body.active, revision: body.expectedRevision + 1 };
  return send(200, { position: state.discoveryPosition });
}
