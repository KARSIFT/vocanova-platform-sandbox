import type { ComponentProps } from "react";
import Link from "./link";
import styles from "./controls.module.css";

export type ActionVariant = "primary" | "secondary" | "danger" | "text";

/** Shared action styling; retain native elements for specialized interactions. */
export function actionClasses(
  variant: ActionVariant = "primary",
  className?: string,
): string {
  return [
    "vn-btn",
    variant === "primary" || variant === "secondary"
      ? `vn-btn-${variant}`
      : styles[variant],
    styles.action,
    className,
  ]
    .filter(Boolean)
    .join(" ");
}

export function Button({
  variant = "primary",
  type = "button",
  className,
  ...props
}: ComponentProps<"button"> & { variant?: ActionVariant }) {
  return (
    <button
      {...props}
      type={type}
      className={actionClasses(variant, className)}
    />
  );
}

/** A navigation action stays a real link, including modifier-click behavior. */
export function ActionLink({
  variant = "primary",
  className,
  ...props
}: ComponentProps<typeof Link> & { variant?: ActionVariant }) {
  return <Link {...props} className={actionClasses(variant, className)} />;
}

export function TextInput({ className, ...props }: ComponentProps<"input">) {
  return (
    <input
      {...props}
      className={[styles.field, className].filter(Boolean).join(" ")}
    />
  );
}

export function TextArea({ className, ...props }: ComponentProps<"textarea">) {
  return (
    <textarea
      {...props}
      className={[styles.field, styles.textarea, className]
        .filter(Boolean)
        .join(" ")}
    />
  );
}

export function Select({ className, ...props }: ComponentProps<"select">) {
  return (
    <select
      {...props}
      className={[styles.field, className].filter(Boolean).join(" ")}
    />
  );
}
