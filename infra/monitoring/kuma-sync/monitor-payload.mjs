import { buildDescription } from "./monitor-metadata.mjs";

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
  "conditions",
  "notificationIDList",
  "upsideDown",
  "maxredirects",
  "ignoreTls",
];

export function inventoryEntryToDesiredMonitor(entry, ownershipMarker) {
  const keyword =
    entry.expected_body === null || entry.expected_body === undefined
      ? ""
      : String(entry.expected_body);

  return {
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
    conditions: [],
    notificationIDList: {},
    upsideDown: false,
    maxredirects: 10,
    ignoreTls: false,
  };
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
    String(actual?.description ?? "") === String(desired.description ?? "")
  );
}

export function snapshotMonitorForRollback(monitor) {
  const snapshot = { id: monitor.id };
  for (const field of EDITABLE_MONITOR_FIELDS) {
    if (field in monitor) {
      snapshot[field] = monitor[field];
    }
  }
  if (!("conditions" in snapshot)) {
    snapshot.conditions = [];
  }
  return snapshot;
}

export function buildEditPayload(desired, kumaId) {
  const payload = snapshotMonitorForRollback(desired);
  payload.id = kumaId;
  return payload;
}
