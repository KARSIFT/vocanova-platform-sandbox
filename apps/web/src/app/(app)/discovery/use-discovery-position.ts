"use client";
import { useEffect, useRef, useState } from "react";
import { ApiResponseError } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { DiscoveryPositionWriter } from "./discovery-position-writer";

export function useDiscoveryPosition({
  catalogVersion,
  band,
  rank,
  active,
  revision,
  enabled,
}: {
  catalogVersion: string;
  band: number;
  rank: number | undefined;
  active: boolean;
  revision: number | null;
  enabled: boolean;
}) {
  const [status, setStatus] = useState<
    "saving" | "saved" | "error" | "conflict"
  >("saving");
  const writer = useRef<DiscoveryPositionWriter | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    const client = createApiClient();
    const queue = new DiscoveryPositionWriter(revision, {
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
      conflict: (error) =>
        error instanceof ApiResponseError && error.status === 409,
    });
    writer.current = queue;
    return () => {
      queue.dispose();
      controller.abort();
      writer.current = null;
    };
  }, [revision]);
  useEffect(() => {
    if (enabled && rank !== undefined)
      writer.current?.select({ catalogVersion, band, rank, active });
  }, [catalogVersion, band, rank, active, enabled, revision]);
  return { status, retry: () => void writer.current?.retry() };
}
