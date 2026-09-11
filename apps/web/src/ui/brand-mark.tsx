import type { ComponentPropsWithoutRef } from "react";

export function BrandMark({
  compact = false,
  className,
  ...props
}: ComponentPropsWithoutRef<"span"> & { compact?: boolean }) {
  return (
    <span
      {...props}
      className={`inline-flex items-center gap-[var(--spacing-sm)] ${className ?? ""}`}
    >
      <span
        aria-hidden="true"
        className="grid size-9 place-items-center rounded-[0.7rem] bg-primary-700 text-sm font-bold tracking-[-0.08em] text-white shadow-[0_5px_14px_rgb(30_58_138_/_0.2)]"
      >
        V
      </span>
      {!compact ? (
        <span className="leading-tight">
          <span className="block text-lg font-bold tracking-[-0.045em] text-neutral-900">
            VocaNova
          </span>
          <span className="block text-xs font-medium text-neutral-600">
            practical English
          </span>
        </span>
      ) : null}
    </span>
  );
}
