import { ActionLink } from "@/ui/controls";
import type { Metadata } from "next";
import Link from "@/ui/link";

import { BrandMark } from "@/ui/brand-mark";
import { LearningPreview } from "@/ui/learning-preview";
import styles from "./public-page.module.css";

export const metadata: Metadata = {
  title: "Vocanova — English for the moments that matter",
  description:
    "Build practical English vocabulary through real-life journeys, smart review, and supportive sentence feedback.",
};

const benefits = [
  ["Discover", "Find useful words in the places and conversations you know."],
  [
    "Remember",
    "Review words in short sessions, when they are ready to return.",
  ],
  ["Use", "Write a sentence and get feedback that helps the word stick."],
] as const;

export default function Page() {
  return (
    <main className={styles.canvas}>
      <div className={styles.container}>
        <header className={styles.header}>
          <Link href="/" aria-label="VocaNova home" className={styles.brand}>
            <BrandMark />
          </Link>
          <ActionLink variant="secondary" href="/login" className="text-sm">
            Sign in
          </ActionLink>
        </header>

        <section className={styles.hero} aria-labelledby="landing-heading">
          <div className={styles.heroCopy}>
            <h1 id="landing-heading" className={styles.heroTitle}>
              Learn the words you will actually use.
            </h1>
            <p className={styles.heroDescription}>
              VocaNova gives each useful word a real setting, then helps you
              return to it, use it, and make it yours.
            </p>
            <div className={styles.heroActions}>
              <ActionLink href="/login">Start learning</ActionLink>
              <a href="#how-it-works" className={styles.textLink}>
                How it works
              </a>
            </div>
          </div>

          <LearningPreview />
        </section>

        <section
          id="how-it-works"
          aria-labelledby="how-it-works-heading"
          className={styles.howItWorks}
        >
          <h2 id="how-it-works-heading" className={styles.sectionTitle}>
            A little practice, with a clear purpose.
          </h2>
          <ol className={styles.learningLoop} role="list">
            {benefits.map(([label, copy]) => (
              <li key={label}>
                <h3>{label}</h3>
                <p>{copy}</p>
              </li>
            ))}
          </ol>
        </section>
        <footer className={styles.footer}>
          <Link href="/help" className={styles.textLink}>
            Help and support
          </Link>
        </footer>
      </div>
    </main>
  );
}
