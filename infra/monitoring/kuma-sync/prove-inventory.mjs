import { CANONICAL_AVAILABILITY_MONITOR_IDS } from "../validate-inventory.mjs";
import {
  inventoryEntryToDesiredMonitor,
  monitorsMatch,
} from "./monitor-payload.mjs";
import {
  isRepoManagedMonitor,
  parseManagedDescription,
} from "./monitor-metadata.mjs";
import { monitorUrlsEqual } from "./url-compare.mjs";

function adoptionMatches(monitor, adoption) {
  return (
    String(monitor?.name ?? "") === String(adoption.match_name) &&
    monitorUrlsEqual(monitor?.url, adoption.match_url)
  );
}

function indexRemoteMonitors(remoteMonitors, ownershipMarker) {
  const managedByRepoId = new Map();
  const unmanaged = [];

  for (const [rawId, monitor] of Object.entries(remoteMonitors ?? {})) {
    const kumaId = Number(rawId);
    const parsed = parseManagedDescription(monitor?.description);

    if (isRepoManagedMonitor(monitor, ownershipMarker) && parsed?.monitorId) {
      if (managedByRepoId.has(parsed.monitorId)) {
        throw new Error(
          `duplicate live repository-managed monitor id: ${parsed.monitorId}`,
        );
      }
      managedByRepoId.set(parsed.monitorId, { kumaId, monitor });
      continue;
    }

    unmanaged.push({ kumaId, monitor });
  }

  return { managedByRepoId, unmanaged };
}

function findInventoryMonitor(
  entry,
  { managedByRepoId, unmanaged },
  ownershipMarker,
) {
  const managed = managedByRepoId.get(entry.id);
  if (managed) {
    return managed;
  }

  if (entry.adoption) {
    const candidates = unmanaged.filter(
      ({ monitor }) =>
        !isRepoManagedMonitor(monitor, ownershipMarker) &&
        adoptionMatches(monitor, entry.adoption),
    );
    if (candidates.length === 1) {
      return candidates[0];
    }
  }

  return null;
}

function redactedMonitorFields(monitor) {
  return {
    name: monitor?.name ?? "",
    url: monitor?.url ?? "",
    interval: monitor?.interval ?? null,
    timeout: monitor?.timeout ?? null,
    maxretries: monitor?.maxretries ?? null,
    accepted_statuscodes: Array.isArray(monitor?.accepted_statuscodes)
      ? monitor.accepted_statuscodes.map(String)
      : [],
    keyword: monitor?.keyword ? "[present]" : "",
    active: monitor?.active !== false,
  };
}

export function proveKumaInventory({
  monitorsDocument,
  remoteMonitors,
  ownershipMarker,
}) {
  const inventoryMonitors = monitorsDocument.availability_monitors ?? [];
  const { managedByRepoId, unmanaged } = indexRemoteMonitors(
    remoteMonitors,
    ownershipMarker,
  );

  const results = [];
  let failures = 0;

  for (const inventoryId of CANONICAL_AVAILABILITY_MONITOR_IDS) {
    const entry = inventoryMonitors.find((item) => item.id === inventoryId);
    if (!entry) {
      results.push({
        inventoryId,
        status: "fail",
        reason: "missing from canonical inventory document",
      });
      failures += 1;
      continue;
    }

    const located = findInventoryMonitor(
      entry,
      { managedByRepoId, unmanaged },
      ownershipMarker,
    );

    if (!located) {
      results.push({
        inventoryId,
        status: "fail",
        reason: "not present in live Kuma monitor list",
        expectedName: entry.name,
        expectedUrl: entry.url,
      });
      failures += 1;
      continue;
    }

    // Preserve-by-default fields (currently notificationIDList) are resolved
    // against the located live monitor. Building desired before lookup would
    // falsely report drift for a correctly preserved notification binding.
    const desired = inventoryEntryToDesiredMonitor(entry, ownershipMarker, {
      remoteMonitor: located.monitor,
    });

    const managed = isRepoManagedMonitor(located.monitor, ownershipMarker);
    if (!monitorsMatch(located.monitor, desired)) {
      results.push({
        inventoryId,
        status: "fail",
        reason: managed
          ? "live monitor metadata does not match inventory"
          : "adopted monitor metadata does not match inventory (run sync to apply ownership marker)",
        kumaId: located.kumaId,
        liveName: located.monitor?.name ?? "",
        liveUrl: located.monitor?.url ?? "",
        repoManaged: managed,
        metadata: redactedMonitorFields(located.monitor),
      });
      failures += 1;
      continue;
    }

    results.push({
      inventoryId,
      status: "pass",
      kumaId: located.kumaId,
      repoManaged: managed,
      metadata: redactedMonitorFields(located.monitor),
    });
  }

  return { results, failures, managedCount: managedByRepoId.size };
}

export function formatProofSummary({ results, failures, managedCount }) {
  const lines = [
    "VOC-086 Kuma inventory proof (read-only Socket.IO)",
    `Managed monitors in Kuma: ${managedCount}`,
    "",
  ];

  for (const result of results) {
    if (result.status === "pass") {
      const meta = result.metadata ?? {};
      lines.push(
        `PASS: ${result.inventoryId} (kuma_id=${result.kumaId}, name=${meta.name}, url=${meta.url}, interval=${meta.interval}, timeout=${meta.timeout}, retries=${meta.maxretries}, statuscodes=${(meta.accepted_statuscodes ?? []).join("|")}, repo_managed=${result.repoManaged}, active=${meta.active})`,
      );
      continue;
    }

    let line = `FAIL: ${result.inventoryId} — ${result.reason}`;
    if (result.expectedName) {
      line += ` (expected name=${result.expectedName})`;
    }
    if (result.expectedUrl) {
      line += ` (expected url=${result.expectedUrl})`;
    }
    if (result.liveName) {
      line += ` (live name=${result.liveName}, url=${result.liveUrl})`;
    }
    lines.push(line);
  }

  lines.push(
    failures === 0
      ? "\nAll canonical availability monitors match inventory."
      : `\n${failures} monitor(s) failed inventory proof.`,
  );

  return lines.join("\n");
}
