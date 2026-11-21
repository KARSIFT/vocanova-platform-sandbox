"use client";

import Link from "@/ui/link";
import { usePathname } from "next/navigation";

import { isPrimaryNavItemActive } from "./bottom-nav-state";
import { HomeIcon, JourneyIcon, ProgressIcon } from "./nav-icons";
import { getSessionExit } from "./session-route";

const NAV_ITEMS = [
  { href: "/home", label: "Home", icon: HomeIcon },
  { href: "/discover", label: "Journey", icon: JourneyIcon },
  { href: "/progress", label: "Progress", icon: ProgressIcon },
] as const;

export function BottomNav() {
  const pathname = usePathname();
  if (getSessionExit(pathname)) return null;

  return (
    <nav
      aria-label="Primary"
      className="fixed inset-x-0 bottom-0 z-20 border-t border-neutral-200 bg-white pb-[env(safe-area-inset-bottom)] lg:hidden"
    >
      <div className="mx-auto flex min-h-16 w-full max-w-[48rem] px-2 py-1">
        {NAV_ITEMS.map((item) => {
          const isActive = isPrimaryNavItemActive(pathname, item.href);
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
              // Primary destinations are worth a viewport prefetch.
              prefetch="auto"
              aria-current={isActive ? "page" : undefined}
              className={`group flex min-h-11 min-w-11 flex-1 flex-col items-center justify-center gap-0.5 rounded-xl text-center text-xs whitespace-nowrap transition-colors ${
                isActive
                  ? "font-semibold text-primary-800"
                  : "font-medium text-neutral-600 hover:text-neutral-900"
              }`}
            >
              <span
                aria-hidden="true"
                className={`flex h-7 w-14 items-center justify-center rounded-full transition-colors duration-[var(--duration-fast)] ${
                  isActive
                    ? "bg-primary-100 text-primary-800"
                    : "group-hover:bg-neutral-100"
                }`}
              >
                <Icon />
              </span>
              {item.label}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
