import assert from "node:assert/strict";
import { afterEach, it, mock } from "node:test";

import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient } from "../../src/lib/api-server";
import { fetchWithServerReadDeadline } from "../../src/lib/server-read-fetch";

const originalFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = originalFetch;
  mock.restoreAll();
});

function controlDeadline() {
  let deadline: AbortController;
  const timeout = mock.method(
    AbortSignal,
    "timeout",
    (milliseconds: number) => {
      assert.equal(milliseconds, 8_000);
      deadline = new AbortController();
      return deadline.signal;
    },
  );
  return {
    timeout,
    expire: () =>
      deadline.abort(new DOMException("Read expired", "TimeoutError")),
  };
}

// A fetch that never receives headers unless its actual supplied signal aborts.
// The injected deadline is controlled explicitly; no wall-clock race is used.
function pendingFetch() {
  let signal: AbortSignal | null | undefined;
  let settle!: (response: Response) => void;
  globalThis.fetch = ((_input, init) => {
    signal = init?.signal;
    return new Promise<Response>((resolve, reject) => {
      settle = resolve;
      if (signal?.aborted) reject(signal.reason);
      else
        signal?.addEventListener("abort", () => reject(signal?.reason), {
          once: true,
        });
    });
  }) as typeof fetch;
  return {
    signal: () => signal,
    release: () => settle(Response.json({ onboardingStatus: "completed" })),
  };
}

it("bounds a stalled server GET and allows a fresh read after a timeout", async () => {
  const deadline = controlDeadline();
  const pending = pendingFetch();
  const client = await createServerApiClient();
  const read = client.getCurrentUser();
  const rejection = assert.rejects(read, { name: "TimeoutError" });
  try {
    assert.equal(deadline.timeout.mock.callCount(), 1);
    assert.equal(pending.signal()?.aborted, false);
    deadline.expire();
    await rejection;
  } finally {
    // Keep the pre-fix regression run from leaving a pending promise behind.
    pending.release();
    await rejection.catch(() => {});
  }
  globalThis.fetch = async (_input, init) => {
    assert.equal(init?.signal?.aborted, false);
    assert.notEqual(init?.signal, pending.signal());
    return Response.json({ id: "recovered" });
  };
  const recovered = await client.getCurrentUser();
  assert.equal(recovered.data.id, "recovered");
});

it("keeps the deadline active while the response body is still arriving", async () => {
  const deadline = controlDeadline();
  let bodyStarted!: () => void;
  const readingBody = new Promise<void>((resolve) => {
    bodyStarted = resolve;
  });
  globalThis.fetch = (async (_input, init) =>
    new Response(
      new ReadableStream({
        start(controller) {
          controller.enqueue(new TextEncoder().encode('{"id":'));
          const signal = init?.signal;
          signal?.addEventListener(
            "abort",
            () => controller.error(signal.reason),
            { once: true },
          );
          bodyStarted();
        },
      }),
      { headers: { "Content-Type": "application/json" } },
    )) as typeof fetch;
  const client = await createServerApiClient();
  const read = client.getCurrentUser();
  const rejection = assert.rejects(read, { name: "TimeoutError" });
  await readingBody;
  deadline.expire();
  await rejection;
});

it("preserves a caller abort reason instead of replacing its signal", async () => {
  controlDeadline();
  pendingFetch();
  const caller = new AbortController();
  const client = await createServerApiClient();
  const read = client.getCurrentUser({ signal: caller.signal });
  const reason = new DOMException("Caller cancelled", "AbortError");
  const rejection = assert.rejects(read, (error) => error === reason);
  caller.abort(reason);
  await rejection;
});

it("still times out when the caller supplies an unexpired signal", async () => {
  const deadline = controlDeadline();
  pendingFetch();
  const caller = new AbortController();
  const client = await createServerApiClient();
  const read = client.getCurrentUser({ signal: caller.signal });
  const rejection = assert.rejects(read, { name: "TimeoutError" });
  deadline.expire();
  await rejection;
  assert.equal(caller.signal.aborted, false);
});

it("leaves server mutations and their caller signal unchanged", async () => {
  const deadline = controlDeadline();
  const caller = new AbortController();
  let calls = 0;
  globalThis.fetch = (async (_input, init) => {
    calls += 1;
    assert.equal(init?.method, "POST");
    assert.equal(init?.signal, caller.signal);
    return new Response(null, { status: 204 });
  }) as typeof fetch;
  const client = await createServerApiClient();
  await client.logout({ signal: caller.signal });
  assert.equal(calls, 1);
  assert.equal(deadline.timeout.mock.callCount(), 0);
});

it("keeps an actual 401 distinguishable from a timeout", async () => {
  controlDeadline();
  globalThis.fetch = async () => new Response(null, { status: 401 });
  const client = await createServerApiClient();
  await assert.rejects(
    client.getCurrentUser(),
    (error) => error instanceof ApiResponseError && error.status === 401,
  );
});

it("bounds HEAD reads and preserves Request-object cancellation", async () => {
  const deadline = controlDeadline();
  pendingFetch();
  const caller = new AbortController();
  const request = new Request("http://api.internal/status", {
    method: "HEAD",
    signal: caller.signal,
  });
  const read = fetchWithServerReadDeadline(request);
  const reason = new DOMException("Caller cancelled HEAD", "AbortError");
  const rejection = assert.rejects(read, (error) => error === reason);
  caller.abort(reason);
  await rejection;
  assert.equal(deadline.timeout.mock.callCount(), 1);
});

it("immediately preserves an already-aborted caller signal", async () => {
  controlDeadline();
  pendingFetch();
  const reason = new DOMException("Already cancelled", "AbortError");
  await assert.rejects(
    fetchWithServerReadDeadline("http://api.internal/status", {
      signal: AbortSignal.abort(reason),
    }),
    (error) => error === reason,
  );
});

it("honours a method override on a Request without timing out a mutation", async () => {
  const deadline = controlDeadline();
  globalThis.fetch = async () => new Response(null, { status: 204 });
  await fetchWithServerReadDeadline(new Request("http://api.internal/status"), {
    method: "PATCH",
  });
  assert.equal(deadline.timeout.mock.callCount(), 0);
});
