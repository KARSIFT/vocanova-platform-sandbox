export function getSessionExit(
  pathname: string,
): { href: string; label: string } | undefined {
  if (pathname.startsWith("/learn/"))
    return { href: "/discover", label: "Back to Journey" };
  if (pathname.startsWith("/practice/session/"))
    return { href: "/practice", label: "Back to Practice" };
  if (pathname.startsWith("/stories/session/"))
    return { href: "/stories", label: "Back to Stories" };
  return undefined;
}
