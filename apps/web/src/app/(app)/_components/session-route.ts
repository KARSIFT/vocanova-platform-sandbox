export function getSessionExit(
  pathname: string,
): { href: string; label: string } | undefined {
  if (
    pathname === "/review" ||
    pathname === "/reviews" ||
    pathname.startsWith("/review/") ||
    pathname.startsWith("/reviews/")
  )
    return { href: "/home", label: "Back to Home" };
  if (pathname === "/discovery")
    return { href: "/knowledge-map", label: "Back to word map" };
  if (pathname.startsWith("/learn/"))
    return { href: "/courses", label: "Back to courses" };
  if (pathname.startsWith("/practice/session/"))
    return { href: "/practice", label: "Back to Practice" };
  if (pathname.startsWith("/stories/session/"))
    return { href: "/stories", label: "Back to Stories" };
  return undefined;
}
