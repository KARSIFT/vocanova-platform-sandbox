import type { ComponentPropsWithoutRef, ReactNode } from "react";

function joinClasses(...classes: Array<string | undefined>) {
  return classes.filter(Boolean).join(" ");
}

export function PageContainer({
  children,
  className,
  ...props
}: ComponentPropsWithoutRef<"div">) {
  return (
    <div
      className={joinClasses(
        "mx-auto w-full max-w-[48rem] px-[var(--spacing-md)] py-[var(--spacing-lg)] sm:px-[var(--spacing-xl)] sm:py-[var(--spacing-xl)]",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function Surface({
  children,
  className,
  tone = "default",
  ...props
}: ComponentPropsWithoutRef<"section"> & {
  tone?: "default" | "primary" | "secondary";
}) {
  return (
    <section
      className={joinClasses(
        "rounded-[var(--radius-lg)] border p-[var(--spacing-md)] shadow-sm sm:p-[var(--spacing-lg)]",
        tone === "primary"
          ? "border-primary-200 bg-primary-50"
          : tone === "secondary"
            ? "border-secondary-200 bg-secondary-50"
            : "border-neutral-200 bg-white",
        className,
      )}
      {...props}
    >
      {children}
    </section>
  );
}

export function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <p className="text-sm font-semibold tracking-[0.08em] text-primary-700 uppercase">
      {children}
    </p>
  );
}
