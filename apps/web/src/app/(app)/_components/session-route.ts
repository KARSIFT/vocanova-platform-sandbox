/** Our own location consumers also observe writes made before router effects. */
export const SESSION_LOCATION_CHANGED = "vocanova:session-location-changed";

export function replaceSessionLocation(href: string) {
  // Treat framework/browser state as opaque and preserve its exact identity.
  window.history.replaceState(window.history.state, "", href);
  window.dispatchEvent(new Event(SESSION_LOCATION_CHANGED));
}

export function getSessionExit(
  pathname: string,
  search: string | URLSearchParams = "",
): { href: string; label: string } | undefined {
  if (
    pathname === "/review" ||
    pathname === "/reviews" ||
    pathname.startsWith("/review/") ||
    pathname.startsWith("/reviews/")
  )
    return { href: "/home", label: "Back to Home" };
  if (pathname === "/discovery") {
    const query =
      typeof search === "string" ? new URLSearchParams(search) : search;
    const values = query.getAll("band");
    const band =
      values.length === 1 && /^(?:[1-9]|[1-9][0-9]|100)$/.test(values[0] ?? "")
        ? values[0]
        : "1";
    return {
      href: `/discover?band=${band}#range-words`,
      label: "Back to Journey",
    };
  }
  if (pathname.startsWith("/learn/"))
    return { href: "/courses", label: "Back to courses" };
  if (pathname.startsWith("/practice/session/"))
    return { href: "/practice", label: "Back to Practice" };
  if (pathname.startsWith("/stories/session/"))
    return { href: "/stories", label: "Back to Stories" };
  return undefined;
}
