"use client";
import { useEffect, useRef, useState } from "react";
import { ApiResponseError } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { DiscoveryPositionWriter } from "./discovery-position-writer";

export function useDiscoveryPosition({
  accountId,
  catalogVersion,
  band,
  rank,
  active,
  revision,
  enabled,
}: {
  accountId: string | null;
  catalogVersion: string;
  band: number;
  rank: number | undefined;
  active: boolean;
  revision: number | null;
  enabled: boolean;
}) {
  const [status, setStatus] = useState<
    "saving" | "saved" | "error" | "conflict" | "account-changed"
  >("saving");
  const writer = useRef<DiscoveryPositionWriter | null>(null);
  useEffect(() => {
    // Never acquire ownership in a later client read: the cookies may now be
    // another account's. A verified rendered owner is required to save.
    if (!accountId) {
      setStatus("error");
      return;
    }
    setStatus("saving");
    const controller = new AbortController();
    const client = createApiClient();
    const queue = new DiscoveryPositionWriter(accountId, revision, {
      read: async () =>
        (await client.getDiscoveryPosition({ signal: controller.signal })).data
          .position?.revision ?? 0,
      write: async (value) => {
        const csrf = await getOrRefreshCSRFToken();
        if (controller.signal.aborted) throw new Error("Discovery closed");
        if (!csrf) throw new Error("Session not ready");
        const response = await client.putDiscoveryPosition(value, {
          signal: controller.signal,
          headers: { "X-CSRF-Token": csrf },
        });
        if (!response.data.position) throw new Error("Position not confirmed");
        return response.data.position.revision;
      },
      status: setStatus,
      accountChanged: (error) =>
        error instanceof ApiResponseError &&
        error.status === 403 &&
        error.body?.detail === "Discovery account changed; reload discovery",
      conflict: (error) =>
        error instanceof ApiResponseError && error.status === 409,
    });
    writer.current = queue;
    return () => {
      queue.dispose();
      controller.abort();
      writer.current = null;
    };
  }, [accountId, revision]);
  useEffect(() => {
    if (enabled && rank !== undefined)
      writer.current?.select({ catalogVersion, band, rank, active });
  }, [catalogVersion, band, rank, active, enabled, revision, accountId]);
  return { status, retry: () => void writer.current?.retry() };
}
