import type { ComponentProps } from "react";
import styles from "./status-message.module.css";

/** State owners keep the message mounted and decide when its text changes.
 * Use live={false} for static guidance or inside an existing live region.
 */
export function StatusMessage({
  as: Element = "div",
  tone = "info",
  live = true,
  role,
  className,
  ...props
}: ComponentProps<"div"> & {
  as?: "div" | "p";
  tone?: "info" | "success" | "error";
  live?: boolean;
}) {
  return (
    <Element
      {...props}
      role={
        role ?? (live ? (tone === "error" ? "alert" : "status") : undefined)
      }
      className={[styles.message, styles[tone], className]
        .filter(Boolean)
        .join(" ")}
    />
  );
}
