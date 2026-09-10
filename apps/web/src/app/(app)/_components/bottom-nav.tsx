"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { isPrimaryNavItemActive } from "./bottom-nav-state";

const NAV_ITEMS = [
  { href: "/home", label: "Home", icon: HomeIcon },
  { href: "/discover", label: "Journey", icon: JourneyIcon },
  { href: "/progress", label: "Progress", icon: ProgressIcon },
] as const;

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="Primary"
      className="fixed inset-x-0 bottom-0 z-20 border-t border-neutral-200/90 bg-white/95 pb-[env(safe-area-inset-bottom)] backdrop-blur"
    >
      <div className="mx-auto flex h-16 w-full max-w-[48rem]">
        {NAV_ITEMS.map((item) => {
          const isActive = isPrimaryNavItemActive(pathname, item.href);
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isActive ? "page" : undefined}
              className={`relative flex min-h-11 min-w-11 flex-1 flex-col items-center justify-center gap-0.5 rounded-xl text-xs transition-colors ${
                isActive
                  ? "font-semibold text-primary-800"
                  : "font-medium text-neutral-500 hover:text-neutral-800"
              }`}
            >
              <Icon />
              {item.label}
              {isActive ? (
                <span
                  aria-hidden="true"
                  className="absolute bottom-1 h-1 w-1 rounded-full bg-primary-700"
                />
              ) : null}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}

function HomeIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-5 w-5 fill-none stroke-current stroke-[1.9]"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="m4 10.5 8-6.25 8 6.25v8.25a1 1 0 0 1-1 1h-4.25v-5.5h-5.5v5.5H5a1 1 0 0 1-1-1V10.5Z"
      />
    </svg>
  );
}

function JourneyIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-5 w-5 fill-none stroke-current stroke-[1.9]"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M5 18.5c3.5-1.5 5.75-4.1 7-8.1 1.1 2.45 3.35 4.2 7 4.95M5 6.25h14M5 11.5h4"
      />
      <circle cx="5" cy="18.5" r="1.35" fill="currentColor" stroke="none" />
    </svg>
  );
}

function ProgressIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      className="h-5 w-5 fill-none stroke-current stroke-[1.9]"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M5 19.25V12m7 7.25V5m7 14.25v-9"
      />
    </svg>
  );
}
