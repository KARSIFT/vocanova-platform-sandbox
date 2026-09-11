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
        "mx-auto w-full max-w-[48rem] px-[var(--spacing-md)] py-[var(--spacing-lg)] sm:px-[var(--spacing-xl)] sm:py-[2.5rem]",
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
        "rounded-[0.9rem] border p-[var(--spacing-md)] shadow-[0_1px_2px_rgb(15_23_42_/_0.05)] sm:p-[var(--spacing-lg)]",
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
  return <p className="text-sm font-semibold text-primary-700">{children}</p>;
}
