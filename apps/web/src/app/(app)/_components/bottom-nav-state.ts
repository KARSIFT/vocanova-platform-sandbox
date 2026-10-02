export function isPrimaryNavItemActive(
  pathname: string,
  href: string,
): boolean {
  return (
    pathname === href ||
    (href === "/progress" && pathname.startsWith("/progress/")) ||
    (href === "/discover" &&
      (pathname.startsWith("/discover/") ||
        pathname === "/plan" ||
        pathname === "/vocabulary" ||
        pathname.startsWith("/vocabulary/") ||
        pathname === "/practice" ||
        pathname.startsWith("/practice/") ||
        pathname.startsWith("/learn/") ||
        pathname === "/words" ||
        pathname.startsWith("/words/")))
  );
}
