export function isPrimaryNavItemActive(
  pathname: string,
  href: string,
): boolean {
  return (
    pathname === href ||
    (href === "/discover" && pathname.startsWith("/discover/"))
  );
}
