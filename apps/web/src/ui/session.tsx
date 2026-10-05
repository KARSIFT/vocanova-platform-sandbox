import type { ComponentPropsWithoutRef } from "react";

/** A stable task action area; players reserve space so it never covers content. */
export function SessionActions({
  children,
  className = "",
  ...props
}: ComponentPropsWithoutRef<"div">) {
  return (
    <div
      className={`fixed inset-x-0 bottom-0 z-20 border-t border-neutral-200 bg-white/95 px-4 pb-[calc(1rem+env(safe-area-inset-bottom))] pt-3 backdrop-blur ${className}`}
      {...props}
    >
      <div className="mx-auto flex w-full max-w-[48rem] flex-wrap items-center justify-between gap-3">
        {children}
      </div>
    </div>
  );
}
