"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const NAV_ITEMS = [
  { href: "/home", label: "Home" },
  { href: "/discover", label: "Journey" },
  { href: "/progress", label: "Progress" },
] as const;

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="Primary"
      className="fixed inset-x-0 bottom-0 flex h-16 w-full border-t border-neutral-200 bg-white"
    >
      {NAV_ITEMS.map((item) => {
        const isActive = pathname === item.href;

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? "page" : undefined}
            className={`flex min-h-11 min-w-11 flex-1 flex-col items-center justify-center gap-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-blue-600 ${
              isActive
                ? "font-semibold text-blue-700"
                : "font-normal text-neutral-600"
            }`}
          >
            <span
              aria-hidden="true"
              className={`h-0.5 w-8 rounded-full ${
                isActive ? "bg-blue-700" : "bg-transparent"
              }`}
            />
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
