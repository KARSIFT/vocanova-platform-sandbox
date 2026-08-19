/**
 * Shared monitor URL comparison for adoption matching and unmanaged URL collision.
 * HTTP(S) URLs are equivalent when they differ only by trailing slashes.
 */
export function normalizeMonitorUrl(url) {
  const value = String(url ?? "");
  if (!/^https?:\/\//i.test(value)) {
    return value;
  }
  return value.replace(/\/+$/, "") || value;
}

export function monitorUrlsEqual(left, right) {
  const leftString = String(left ?? "");
  const rightString = String(right ?? "");
  if (leftString === rightString) {
    return true;
  }
  return normalizeMonitorUrl(leftString) === normalizeMonitorUrl(rightString);
}
