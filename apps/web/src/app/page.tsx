import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Vocanova — English for the moments that matter",
  description:
    "Build practical English vocabulary through real-life journeys, smart review, and supportive sentence feedback.",
};

const benefits = [
  {
    label: "Learn in context",
    copy: "Explore useful words for airports, restaurants, work, university, and everyday conversations.",
    marker: "01",
  },
  {
    label: "Remember for longer",
    copy: "Short daily reviews return each word when it is most useful to practise it again.",
    marker: "02",
  },
  {
    label: "Turn words into speech",
    copy: "Write your own sentence and get concise, encouraging feedback focused on the word you chose.",
    marker: "03",
  },
] as const;

export default function Page() {
  return (
    <main className="min-h-screen overflow-hidden bg-[#f7f4ec] text-[#17352b]">
      <div className="mx-auto flex min-h-screen max-w-[78rem] flex-col px-[var(--spacing-md)] sm:px-[var(--spacing-xl)]">
        <header className="flex min-h-20 items-center justify-between border-b border-[#17352b]/15 py-[var(--spacing-md)]">
          <Link
            href="/"
            aria-label="Vocanova home"
            className="rounded-sm text-xl font-semibold tracking-[-0.04em] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[#176b4d]"
          >
            vocanova<span className="text-[#a63f27]">.</span>
          </Link>
          <Link
            href="/login"
            className="inline-flex min-h-11 items-center justify-center rounded-full border border-[#17352b]/30 bg-white/60 px-5 text-sm font-semibold transition-colors hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#176b4d]"
          >
            Sign in
          </Link>
        </header>

        <section className="relative grid flex-1 items-center gap-12 py-16 lg:grid-cols-[1.15fr_0.85fr] lg:py-24">
          <div className="relative z-10 max-w-[45rem]">
            <p className="mb-6 inline-flex min-h-11 items-center rounded-full bg-[#dcebdd] px-4 text-sm font-semibold text-[#245b46]">
              A calmer way to build practical English
            </p>
            <h1 className="text-[clamp(3.25rem,8vw,7.25rem)] leading-[0.9] font-semibold tracking-[-0.075em] text-balance">
              Words for the
              <span className="block text-[#a63f27]">moments that matter.</span>
            </h1>
            <p className="mt-8 max-w-[39rem] text-lg leading-8 text-[#36584c] sm:text-xl">
              Discover English through real situations, remember it with a
              focused daily mission, and practise using every word with helpful
              feedback.
            </p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Link
                href="/login"
                className="inline-flex min-h-12 items-center justify-center rounded-full bg-[#176b4d] px-7 text-base font-semibold text-white shadow-[0_10px_30px_rgba(23,107,77,0.18)] transition-transform hover:-translate-y-0.5 hover:bg-[#12573e] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#176b4d]"
              >
                Start learning
              </Link>
              <a
                href="#how-it-works"
                className="inline-flex min-h-12 items-center justify-center rounded-full px-7 text-base font-semibold text-[#245b46] hover:bg-white/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#176b4d]"
              >
                See how it works
              </a>
            </div>
          </div>

          <div
            aria-hidden="true"
            className="relative mx-auto aspect-square w-full max-w-[28rem]"
          >
            <div className="absolute inset-[8%] rotate-6 rounded-[32%_68%_55%_45%/45%_42%_58%_55%] bg-[#f5b544]" />
            <div className="absolute inset-[19%_12%_14%_20%] -rotate-3 rounded-[2.5rem] border border-white/70 bg-white/85 p-8 shadow-[0_30px_80px_rgba(49,62,47,0.18)] backdrop-blur">
              <p className="text-xs font-bold tracking-[0.18em] text-[#5f685f] uppercase">
                Today&apos;s word
              </p>
              <p className="mt-5 text-4xl font-semibold tracking-[-0.05em]">
                confident
              </p>
              <p className="mt-2 text-sm text-[#637067]">adjective</p>
              <div className="mt-7 h-px bg-[#17352b]/10" />
              <p className="mt-6 text-lg leading-7 text-[#36584c]">
                feeling sure about your ability or decision
              </p>
              <div className="mt-8 flex items-center gap-3 text-sm font-semibold text-[#176b4d]">
                <span className="grid size-9 place-items-center rounded-full bg-[#dcebdd]">
                  ✓
                </span>
                Ready to practise
              </div>
            </div>
            <div className="absolute top-[5%] right-[3%] size-16 rounded-full bg-[#e65f3c]" />
            <div className="absolute bottom-[7%] left-[5%] size-9 rounded-full border-[7px] border-[#176b4d]" />
          </div>
        </section>

        <section
          id="how-it-works"
          aria-labelledby="how-it-works-heading"
          className="border-t border-[#17352b]/15 py-14"
        >
          <div className="grid gap-8 lg:grid-cols-[0.7fr_1.3fr]">
            <div>
              <p className="text-sm font-bold tracking-[0.16em] text-[#a63f27] uppercase">
                The daily loop
              </p>
              <h2
                id="how-it-works-heading"
                className="mt-3 text-3xl font-semibold tracking-[-0.045em] sm:text-4xl"
              >
                Small sessions. Useful progress.
              </h2>
            </div>
            <ol className="grid gap-4 sm:grid-cols-3">
              {benefits.map((benefit) => (
                <li
                  key={benefit.marker}
                  className="rounded-[1.5rem] border border-[#17352b]/10 bg-white/55 p-6"
                >
                  <span className="text-xs font-bold tracking-[0.16em] text-[#a63f27]">
                    {benefit.marker}
                  </span>
                  <h3 className="mt-5 text-lg font-semibold">
                    {benefit.label}
                  </h3>
                  <p className="mt-3 text-sm leading-6 text-[#52675f]">
                    {benefit.copy}
                  </p>
                </li>
              ))}
            </ol>
          </div>
        </section>
      </div>
    </main>
  );
}
