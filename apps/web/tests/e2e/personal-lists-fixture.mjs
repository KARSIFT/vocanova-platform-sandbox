// Synthetic list transport and conflict fixture. This does not prove database durability.
export async function handlePersonalLists({
  req,
  res,
  url,
  cookies,
  state,
  canonicalWords,
  jsonResponse,
  readJsonBody,
  checkCsrf,
  logLine,
}) {
  if (!/^\/api\/v1\/word-lists(?:\/|$)/u.test(url.pathname)) return false;
  const send = (status, body) => {
    jsonResponse(res, status, body);
    return true;
  };
  const gone = () => send(404, { detail: "List not found" });
  const noContent = () => {
    res.writeHead(204);
    res.end();
    return true;
  };
  if (cookies.e2e_lists_auth === "expired")
    return send(401, { detail: "Fixture expired session" });
  if (cookies.e2e_lists === "unavailable")
    return send(503, { detail: "Fixture lists unavailable" });
  state.wordLists ??= new Map();
  state.wordListReceipts ??= new Map();
  state.wordListDeleted ??= new Set();
  const project = (list) => {
    const members = [...list.members].map(([meaningId, addedAt]) => {
      const word = Object.values(canonicalWords).find((item) =>
        item.meanings.some((meaning) => meaning.id === meaningId),
      );
      const meaning = word.meanings.find((item) => item.id === meaningId);
      return {
        meaningId,
        wordId: word.id,
        wordSlug: word.slug,
        wordText: word.text,
        shortDefinition: meaning.shortDefinition,
        partOfSpeech: meaning.partOfSpeech,
        practiceAvailable: ["invite", "confirm", "reschedule"].includes(
          word.slug,
        ),
        addedAt,
      };
    });
    return {
      id: list.id,
      name: list.name,
      revision: list.revision,
      memberCount: members.length,
      usableMemberCount: members.filter((item) => item.practiceAvailable)
        .length,
      createdAt: list.createdAt,
      updatedAt: list.updatedAt,
      members,
    };
  };
  const parts = url.pathname
    .slice("/api/v1/word-lists".length)
    .split("/")
    .filter(Boolean);
  const id = parts[0];
  const list = state.wordLists.get(id);
  if (req.method === "GET") {
    if (parts.length === 0)
      return send(200, {
        items: [...state.wordLists.values()].map((list) => ({
          ...project(list),
          members: undefined,
        })),
      });
    return list && parts.length === 1 ? send(200, project(list)) : gone();
  }
  if (
    !["PUT", "DELETE"].includes(req.method) ||
    !id ||
    (parts.length !== 1 && !(parts.length === 3 && parts[1] === "members"))
  )
    return send(405, { detail: "Method not allowed" });
  if (!checkCsrf(req, cookies, res, logLine)) return true;
  const key = req.headers["idempotency-key"];
  const body = req.method === "PUT" ? await readJsonBody(req) : null;
  const revision =
    req.method === "PUT"
      ? body.expectedRevision
      : Number(url.searchParams.get("expectedRevision"));
  const fingerprint = JSON.stringify([
    req.method,
    url.pathname,
    body,
    req.method === "DELETE" ? revision : null,
  ]);
  if (!key) return send(422, { detail: "Missing mutation identity" });
  const receipt = state.wordListReceipts.get(key);
  if (receipt) {
    if (receipt !== fingerprint)
      return send(409, { detail: "Changed request" });
    if (req.method === "DELETE" && parts.length === 1) return noContent();
    return list ? send(200, project(list)) : gone();
  }
  if (!Number.isSafeInteger(revision) || revision < 0)
    return send(422, { detail: "Invalid revision" });
  if (parts.length === 1 && req.method === "PUT") {
    const name = typeof body.name === "string" ? body.name.trim() : "";
    if (!name || Array.from(name).length > 80 || name.includes("\0"))
      return send(422, { detail: "Use a list name of 1–80 characters" });
    if (state.wordListDeleted.has(id)) return gone();
    if ((list?.revision ?? 0) !== revision)
      return send(409, { detail: "List changed elsewhere" });
    if (!list && state.wordLists.size >= 50)
      return send(422, { detail: "List limit reached" });
    const now = new Date().toISOString();
    const updated = list ?? {
      id,
      members: new Map(),
      revision: 0,
      createdAt: now,
    };
    updated.name = name;
    updated.revision++;
    updated.updatedAt = now;
    state.wordLists.set(id, updated);
    state.wordListReceipts.set(key, fingerprint);
    return send(200, project(updated));
  }
  if (!list) return gone();
  if (revision !== list.revision)
    return send(409, { detail: "List changed elsewhere" });
  if (parts.length === 1) {
    state.wordLists.delete(id);
    state.wordListDeleted.add(id);
    state.wordListReceipts.set(key, fingerprint);
    return noContent();
  }
  const meaningId = decodeURIComponent(parts[2]);
  if (
    !Object.values(canonicalWords).some((item) =>
      item.meanings.some((meaning) => meaning.id === meaningId),
    )
  )
    return gone();
  if (req.method === "PUT") {
    if (!list.members.has(meaningId) && list.members.size >= 500)
      return send(422, { detail: "List meaning limit reached" });
    if (!list.members.has(meaningId))
      list.members.set(meaningId, new Date().toISOString());
  } else list.members.delete(meaningId);
  list.revision++;
  list.updatedAt = new Date().toISOString();
  state.wordListReceipts.set(key, fingerprint);
  return send(200, project(list));
}
