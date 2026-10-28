import type { ReactNode } from "react";
import Link from "@/ui/link";

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
            <span>In context</span>
            <strong>“Could I get this to go?”</strong>
            <p>
              A useful phrase is easier to remember when you can picture the
              moment.
            </p>
          </div>
          <div className="auth-word-card">
            <div className="auth-word-card-top">
              <b>to go</b>
              <em>phrase · ordering at a cafe</em>
            </div>
            <p>Food or drink you take away instead of having it there.</p>
          </div>
          <ol className="auth-loop">
            <li>
              <span>1</span>Discover
            </li>
            <li>
              <span>2</span>Remember
            </li>
            <li>
              <span>3</span>Use
            </li>
          </ol>
          <p className="auth-context-caption">
            A practical English workspace, one useful word at a time.
          </p>
        </div>
      </aside>
      <section className={`auth-panel${wide ? " auth-panel-wide" : ""}`}>
        <Link
          href="/"
          aria-label="VocaNova home"
          className="auth-brand min-h-11"
        >
          <BrandMark />
        </Link>
        <div className="auth-form">{children}</div>
      </section>
    </main>
  );
}
