import Image from "next/image";
import type { ReactNode } from "react";
import Link from "@/ui/link";

import { BrandMark } from "./brand-mark";
import styles from "./auth-shell.module.css";

export function AuthShell({
  children,
  title,
  description,
  wide = false,
  showExample = false,
  linkBrand = true,
}: {
  children: ReactNode;
  title?: string;
  description?: string;
  wide?: boolean;
  showExample?: boolean;
  linkBrand?: boolean;
}) {
  return (
    <main className={styles.canvas}>
      <div className={styles.container}>
        <header className={styles.header}>
          {linkBrand ? (
            <Link href="/" aria-label="VocaNova home" className={styles.brand}>
              <BrandMark />
            </Link>
          ) : (
            <span className={styles.brand}>
              <BrandMark />
            </span>
          )}
        </header>
        <div className={showExample ? styles.withExample : styles.focused}>
          <div className={wide ? styles.widePanel : styles.panel}>
            {title ? (
              <div className={styles.intro}>
                <h1 className="page-title">{title}</h1>
                {description ? (
                  <p className="page-lede">{description}</p>
                ) : null}
              </div>
            ) : null}
            {children}
          </div>
          {showExample ? (
            <figure className={styles.example} aria-hidden="true">
              <div className={styles.exampleHeading}>Ordering food</div>
              <Image
                src="/images/meanings/1e9e55f8-d329-5c52-bc30-7ed77a31a167.webp"
                alt=""
                width={960}
                height={640}
                sizes="(min-width: 960px) 384px, 0px"
                className={styles.exampleImage}
              />
              <figcaption className={styles.exampleCaption}>
                “Could I get this <strong>to go</strong>?”
              </figcaption>
            </figure>
          ) : null}
        </div>
      </div>
    </main>
  );
}

export function AuthDivider({ children }: { children: ReactNode }) {
  return <div className={styles.divider}>{children}</div>;
}
