export function isPrimaryNavItemActive(
  pathname: string,
  href: string,
): boolean {
  return (
    pathname === href ||
    (href === "/progress" && pathname.startsWith("/progress/")) ||
    (href === "/discover" &&
      (pathname.startsWith("/discover/") ||
        pathname === "/words" ||
        pathname.startsWith("/words/")))
  );
}
