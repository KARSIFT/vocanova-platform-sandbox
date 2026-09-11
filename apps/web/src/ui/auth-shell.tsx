import type { ReactNode } from "react";
import Link from "next/link";

import { BrandMark } from "./brand-mark";

export function AuthShell({
  children,
  wide = false,
}: {
  children: ReactNode;
  wide?: boolean;
}) {
  return (
    <main className="auth-shell">
      <aside className="auth-context" aria-hidden="true">
        <div className="auth-context-inner">
          <div className="auth-context-mark">
            <span>in context</span>
            <strong>“Could I get this to go?”</strong>
            <p>
              A useful phrase is easier to remember when you can picture the
              moment.
            </p>
          </div>
          <p className="auth-context-caption">
            A practical English workspace, one useful word at a time.
          </p>
        </div>
      </aside>
      <section className={`auth-panel${wide ? " auth-panel-wide" : ""}`}>
        <Link href="/" aria-label="VocaNova home" className="auth-brand">
          <BrandMark />
        </Link>
        <div className="auth-form">{children}</div>
      </section>
    </main>
  );
}
