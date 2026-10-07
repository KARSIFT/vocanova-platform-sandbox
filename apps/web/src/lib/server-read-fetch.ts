// A stalled secondary read must not hold a server-rendered page indefinitely.
// Allow eight seconds for headers AND body, below the identity gate's ten-second
// deadline. This bounds failure recovery; it does not cache or retry requests.
export const SERVER_READ_DEADLINE_MS = 8_000;

export function fetchWithServerReadDeadline(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  const request = input instanceof Request ? input : undefined;
  const method = (init?.method ?? request?.method ?? "GET").toUpperCase();
  if (method !== "GET" && method !== "HEAD") return fetch(input, init);

  const callerSignal =
    init?.signal === undefined ? request?.signal : init.signal;
  // Native timeout signals have no user-owned timer/listener to leak and remain
  // attached to fetch during response.json(), not just until headers arrive.
  const deadline = AbortSignal.timeout(SERVER_READ_DEADLINE_MS);
  const signal = callerSignal
    ? AbortSignal.any([callerSignal, deadline])
    : deadline;
  return fetch(input, { ...init, signal });
}
