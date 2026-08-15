import {
  inventoryEntryToDesiredMonitor,
  monitorsMatch,
  snapshotMonitorForRollback,
} from "./monitor-payload.mjs";
import {
  isRepoManagedMonitor,
  parseManagedDescription,
} from "./monitor-metadata.mjs";

function adoptionMatches(monitor, adoption) {
  return (
    String(monitor?.name ?? "") === String(adoption.match_name) &&
    String(monitor?.url ?? "") === String(adoption.match_url)
  );
}

export function planSyncOperations({
  inventoryMonitors,
  remoteMonitors,
  ownershipMarker,
}) {
  const errors = [];
  const operations = [];

  const managedByRepoId = new Map();
  const unmanaged = [];

  for (const [rawId, monitor] of Object.entries(remoteMonitors ?? {})) {
    const kumaId = Number(rawId);
    const parsed = parseManagedDescription(monitor?.description);

    if (isRepoManagedMonitor(monitor, ownershipMarker) && parsed?.monitorId) {
      if (managedByRepoId.has(parsed.monitorId)) {
        errors.push(
          `collision: duplicate managed monitor_id ${parsed.monitorId} in Kuma inventory`,
        );
        continue;
      }
      managedByRepoId.set(parsed.monitorId, { kumaId, monitor });
      continue;
    }

    unmanaged.push({ kumaId, monitor });
  }

  for (const entry of inventoryMonitors) {
    const desired = inventoryEntryToDesiredMonitor(entry, ownershipMarker);
    let existing = managedByRepoId.get(entry.id);

    if (!existing && entry.adoption) {
      const candidates = unmanaged.filter(
        ({ monitor }) =>
          !isRepoManagedMonitor(monitor, ownershipMarker) &&
          adoptionMatches(monitor, entry.adoption),
      );

      if (candidates.length > 1) {
        errors.push(
          `collision: multiple adoption candidates for inventory id ${entry.id}`,
        );
        continue;
      }

      if (candidates.length === 1) {
        existing = candidates[0];
      }
    }

    if (!existing) {
      const urlCollision = unmanaged.find(
        ({ monitor }) =>
          !isRepoManagedMonitor(monitor, ownershipMarker) &&
          String(monitor?.url ?? "") === String(entry.url),
      );

      if (urlCollision) {
        errors.push(
          `collision: unmanaged monitor at ${entry.url} blocks create for ${entry.id}; adopt explicitly or remove the manual monitor`,
        );
        continue;
      }

      operations.push({
        type: "create",
        inventoryId: entry.id,
        desired,
      });
      continue;
    }

    if (!monitorsMatch(existing.monitor, desired)) {
      operations.push({
        type: "update",
        inventoryId: entry.id,
        kumaId: existing.kumaId,
        desired,
        previous: snapshotMonitorForRollback(existing.monitor),
      });
    }
  }

  return { operations, errors };
}
