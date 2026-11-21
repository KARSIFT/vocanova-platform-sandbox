import type { ComponentProps, ReactNode } from "react";
import Link from "./link";

function joinClasses(...classes: Array<string | undefined | false>) {
  return classes.filter(Boolean).join(" ");
}

/** Overview pages are broad; word and focused pages keep a reading width. */
const PAGE_WIDTHS = {
  reading: "max-w-[48rem]",
  standard: "max-w-[64rem]",
  wide: "max-w-[72rem]",
} as const;

export function PageContainer({
  children,
  className,
  width = "reading",
  ...props
}: ComponentProps<"div"> & { width?: keyof typeof PAGE_WIDTHS }) {
  return (
    <div
      className={joinClasses(
        "page-container mx-auto w-full px-[var(--spacing-md)] py-[var(--spacing-lg)] sm:px-[var(--spacing-xl)] sm:py-[2.5rem]",
        PAGE_WIDTHS[width],
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

/**
 * The shared page opening: a quiet way back, the page's name and one line of
 * context. Page actions sit beside the title and wrap beneath it on phones.
 */
export function PageHeader({
  title,
  eyebrow,
  lede,
  back,
  actions,
  className,
  children,
}: {
  title: ReactNode;
  eyebrow?: ReactNode;
  lede?: ReactNode;
  back?: { href: string; label: string };
  actions?: ReactNode;
  className?: string;
  children?: ReactNode;
}) {
  return (
    <header className={joinClasses("page-header", className)}>
      {back ? (
        <Link href={back.href} className="page-back">
          <svg aria-hidden="true" viewBox="0 0 24 24">
            <path
              d="m15 5-7 7 7 7"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          {back.label}
        </Link>
      ) : null}
      <div className="page-header-row">
        <div className="page-header-text">
          {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
          <h1 className="page-title">{title}</h1>
          {lede ? <div className="page-lede">{lede}</div> : null}
          {children}
        </div>
        {actions ? <div className="page-header-actions">{actions}</div> : null}
      </div>
    </header>
  );
}

/** A titled region of a page, with an optional trailing link or action. */
export function PageSection({
  title,
  titleId,
  description,
  action,
  children,
  className,
  ...props
}: Omit<ComponentProps<"section">, "title"> & {
  title: ReactNode;
  titleId: string;
  description?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <section
      aria-labelledby={titleId}
      className={joinClasses("page-section", className)}
      {...props}
    >
      <div className="section-head">
        <div className="min-w-0">
          <h2 id={titleId} className="section-title">
            {title}
          </h2>
          {description ? (
            <p className="section-description">{description}</p>
          ) : null}
        </div>
        {action}
      </div>
      {children}
    </section>
  );
}

/**
 * Surfaces follow evidence: the default raised card marks what the page is
 * about, `flat` groups supporting content, and tinted tones carry context.
 */
export function Surface({
  children,
  className,
  tone = "default",
  ...props
}: ComponentProps<"section"> & {
  tone?: "default" | "flat" | "primary" | "secondary";
}) {
  return (
    <section
      className={joinClasses(
        "vn-surface rounded-[1.25rem] border p-[var(--spacing-md)] sm:p-[var(--spacing-lg)]",
        tone === "primary"
          ? "border-primary-200 bg-primary-50"
          : tone === "secondary"
            ? "border-secondary-200 bg-secondary-50"
            : tone === "flat"
              ? "border-neutral-200 bg-white"
              : "border-neutral-200 bg-white shadow-[var(--vn-shadow-card)]",
        className,
      )}
      {...props}
    >
      {children}
    </section>
  );
}

export function Eyebrow({ children }: { children: ReactNode }) {
  return <p className="page-eyebrow">{children}</p>;
}
