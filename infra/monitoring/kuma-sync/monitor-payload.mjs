import { buildDescription } from "./monitor-metadata.mjs";

/**
 * Optional inventory field `notification_id_list` takes explicit ownership of
 * Kuma `notificationIDList`. When omitted, adopt/update preserve the remote
 * monitor's existing bindings (VOC-087-D02 / VOC-087-DEP-03).
 */
export function inventoryOwnsNotificationBindings(entry) {
  return Object.prototype.hasOwnProperty.call(entry, "notification_id_list");
}

export function resolveNotificationIDList(entry, remoteMonitor) {
  if (inventoryOwnsNotificationBindings(entry)) {
    const owned = entry.notification_id_list;
    if (owned === null || typeof owned !== "object" || Array.isArray(owned)) {
      throw new TypeError(
        "notification_id_list must be an object when explicitly configured",
      );
    }
    return structuredClone(owned);
  }

  const remote = remoteMonitor?.notificationIDList;
  if (remote !== null && remote !== undefined && typeof remote === "object") {
    return structuredClone(remote);
  }

  // Kuma's supported add/edit Socket.IO handlers always iterate this mapping.
  // An empty object means no destinations and is required for a safe create;
  // it does not invent a notification binding.
  return {};
}

const EDITABLE_MONITOR_FIELDS = [
  "type",
  "name",
  "url",
  "method",
  "interval",
  "timeout",
  "maxretries",
  "retryInterval",
  "resendInterval",
  "keyword",
  "invertKeyword",
  "accepted_statuscodes",
  "description",
  "active",
  "notificationIDList",
  "upsideDown",
  "maxredirects",
  "ignoreTls",
];

export function inventoryEntryToDesiredMonitor(
  entry,
  ownershipMarker,
  { remoteMonitor = null } = {},
) {
  const keyword =
    entry.expected_body === null || entry.expected_body === undefined
      ? ""
      : String(entry.expected_body);

  const desired = {
    type: "http",
    name: entry.name,
    url: entry.url,
    method: "GET",
    interval: entry.interval_seconds,
    timeout: entry.timeout_seconds,
    maxretries: entry.retries,
    retryInterval: entry.interval_seconds,
    resendInterval: 0,
    keyword,
    invertKeyword: false,
    accepted_statuscodes: [String(entry.expected_status)],
    description: buildDescription({
      monitorId: entry.id,
      severity: entry.severity,
      ownershipMarker,
    }),
    active: true,
    upsideDown: false,
    maxredirects: 10,
    ignoreTls: false,
  };

  desired.notificationIDList = resolveNotificationIDList(entry, remoteMonitor);

  return desired;
}

export function getAcceptedStatusCodes(monitor) {
  if (Array.isArray(monitor?.accepted_statuscodes)) {
    return monitor.accepted_statuscodes.map(String);
  }

  if (typeof monitor?.accepted_statuscodes_json === "string") {
    try {
      const parsed = JSON.parse(monitor.accepted_statuscodes_json);
      if (Array.isArray(parsed)) {
        return parsed.map(String);
      }
    } catch {
      return [];
    }
  }

  return [];
}

export function normalizeKeyword(value) {
  if (value === null || value === undefined) {
    return "";
  }
  return String(value);
}

function normalizeNotificationIDList(value) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) =>
        left.localeCompare(right, "en", { numeric: true }),
      )
      .map(([id, enabled]) => [String(id), Boolean(enabled)]),
  );
}

export function monitorsMatch(actual, desired) {
  return (
    String(actual?.type ?? "") === String(desired.type) &&
    String(actual?.name ?? "") === String(desired.name) &&
    String(actual?.url ?? "") === String(desired.url) &&
    String(actual?.method ?? "GET") === String(desired.method ?? "GET") &&
    Number(actual?.interval) === Number(desired.interval) &&
    Number(actual?.timeout) === Number(desired.timeout) &&
    Number(actual?.maxretries) === Number(desired.maxretries) &&
    normalizeKeyword(actual?.keyword) === normalizeKeyword(desired.keyword) &&
    JSON.stringify(getAcceptedStatusCodes(actual)) ===
      JSON.stringify(desired.accepted_statuscodes.map(String)) &&
    JSON.stringify(normalizeNotificationIDList(actual?.notificationIDList)) ===
      JSON.stringify(normalizeNotificationIDList(desired.notificationIDList)) &&
    String(actual?.description ?? "") === String(desired.description ?? "") &&
    Boolean(actual?.active) === Boolean(desired.active)
  );
}

export function snapshotMonitorForRollback(monitor) {
  const snapshot = { id: monitor.id };
  for (const field of EDITABLE_MONITOR_FIELDS) {
    if (field in monitor) {
      snapshot[field] = monitor[field];
    }
  }
  return snapshot;
}

export function buildEditPayload(desired, kumaId) {
  const payload = snapshotMonitorForRollback(desired);
  payload.id = kumaId;
  return payload;
}
