export function buildDescription({ monitorId, severity, ownershipMarker }) {
  return `${ownershipMarker} monitor_id=${monitorId} severity=${severity}`;
}

export function parseManagedDescription(description) {
  if (typeof description !== "string" || !description.trim()) {
    return null;
  }

  const monitorIdMatch = description.match(/monitor_id=([a-z0-9][a-z0-9.-]*[a-z0-9])/i);
  if (!monitorIdMatch) {
    return null;
  }

  const severityMatch = description.match(/severity=(low|medium|high|critical)/i);

  return {
    monitorId: monitorIdMatch[1],
    severity: severityMatch?.[1]?.toLowerCase() ?? null,
  };
}

export function isRepoManagedMonitor(monitor, ownershipMarker) {
  return (
    typeof monitor?.description === "string" &&
    monitor.description.includes(ownershipMarker) &&
    parseManagedDescription(monitor.description) !== null
  );
}
