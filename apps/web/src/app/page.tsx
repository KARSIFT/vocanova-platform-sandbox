import type { Metadata } from "next";
import Link from "next/link";

import { BrandMark } from "@/ui/brand-mark";
import { LearningPreview } from "@/ui/learning-preview";

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
    <main className="min-h-screen bg-neutral-100 text-neutral-900">
      <div className="mx-auto flex min-h-screen w-full max-w-[76rem] flex-col px-[var(--spacing-md)] sm:px-[var(--spacing-xl)]">
        <header className="flex min-h-20 items-center justify-between border-b border-neutral-200">
          <Link
            href="/"
            aria-label="VocaNova home"
            className="inline-flex min-h-11 items-center gap-[var(--spacing-sm)] rounded-md"
          >
            <BrandMark />
          </Link>
          <Link
            href="/login"
            className="inline-flex min-h-11 items-center justify-center rounded-md border border-neutral-300 bg-white px-5 text-sm font-semibold text-neutral-900 transition-colors hover:border-primary-300 hover:bg-primary-50"
          >
            Sign in
          </Link>
        </header>

        <section className="grid flex-1 items-center gap-10 py-8 lg:grid-cols-[minmax(0,0.95fr)_minmax(24rem,0.8fr)] lg:gap-20 lg:py-24">
          <div className="max-w-[42rem]">
            <p className="text-sm font-semibold text-primary-700">
              Practical English, made personal
            </p>
            <h1 className="mt-4 text-[clamp(2.7rem,5.6vw,5.4rem)] font-semibold leading-[0.98] tracking-[-0.065em] text-balance">
              Learn the words you will actually use.
            </h1>
            <p className="mt-6 max-w-[38rem] text-lg leading-8 text-neutral-700 sm:text-xl">
              VocaNova gives each useful word a real setting, then helps you
              return to it, use it, and make it yours.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                href="/login"
                className="inline-flex min-h-12 items-center justify-center rounded-md bg-primary-700 px-6 text-base font-semibold text-white shadow-[0_7px_16px_rgb(30_58_138_/_0.18)] transition-colors hover:bg-primary-800"
              >
                Start learning
              </Link>
              <a
                href="#how-it-works"
                className="inline-flex min-h-12 items-center justify-center rounded-md px-5 text-base font-semibold text-primary-800 hover:bg-primary-50"
              >
                How it works
              </a>
            </div>
          </div>

          <LearningPreview />
        </section>

        <section
          id="how-it-works"
          aria-labelledby="how-it-works-heading"
          className="border-t border-neutral-200 py-12 lg:py-16"
        >
          <div className="grid gap-8 lg:grid-cols-[minmax(15rem,0.7fr)_minmax(0,1.3fr)] lg:gap-16">
            <div className="max-w-[24rem]">
              <p className="text-sm font-semibold text-primary-700">
                A daily learning loop
              </p>
              <h2
                id="how-it-works-heading"
                className="mt-3 text-3xl font-semibold leading-tight tracking-[-0.045em]"
              >
                A little practice, with a clear purpose.
              </h2>
            </div>
            <ol className="grid gap-0 border-t border-neutral-200 sm:grid-cols-3 sm:border-l">
              {benefits.map(([label, copy]) => (
                <li
                  key={label}
                  className="border-b border-neutral-200 py-5 sm:border-r sm:px-5 sm:py-0"
                >
                  <h3 className="font-semibold text-neutral-900">{label}</h3>
                  <p className="mt-2 text-sm leading-6 text-neutral-600">
                    {copy}
                  </p>
                </li>
              ))}
            </ol>
          </div>
        </section>
        <footer className="border-t border-neutral-200 py-5">
          <Link
            href="/help"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline underline-offset-4"
          >
            Help and support
          </Link>
        </footer>
      </div>
    </main>
  );
}
