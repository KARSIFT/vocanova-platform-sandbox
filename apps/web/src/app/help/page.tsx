import type { Metadata } from "next";
import Link from "next/link";

import { BrandMark } from "@/ui/brand-mark";

export const metadata: Metadata = {
  title: "Help — Vocanova",
  description: "Find your next learning step and get help with Vocanova.",
};

const link =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 underline decoration-primary-300 underline-offset-4 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const question =
  "min-h-12 cursor-pointer content-center py-3 font-semibold text-neutral-900 marker:text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const answer = "space-y-3 pb-5 leading-7 text-neutral-700";

export default function HelpPage() {
  return (
    <main className="min-h-screen bg-neutral-100 text-neutral-900">
      <div className="mx-auto max-w-[52rem] px-4 pb-12 sm:px-8">
        <header className="flex min-h-20 flex-wrap items-center justify-between gap-3 border-b border-neutral-200 py-3">
          <Link
            href="/"
            aria-label="VocaNova home"
            className="inline-flex min-h-11 items-center rounded-md"
          >
            <BrandMark />
          </Link>
          <Link href="/home" className={link}>
            Go to learning
          </Link>
        </header>
        <h1 className="mt-8 text-3xl font-bold tracking-tight">
          How can we help?
        </h1>
        <p className="mt-3 max-w-[40rem] text-lg leading-7 text-neutral-700">
          Find a useful next step, understand your progress, or get in touch.
        </p>

        <section aria-labelledby="learning-help" className="mt-8">
          <h2 id="learning-help" className="text-xl font-bold">
            Learning with Vocanova
          </h2>
          <div className="mt-3 divide-y divide-neutral-200 border-y border-neutral-200">
            <details open>
              <summary className={question}>Where should I start?</summary>
              <div className={answer}>
                <p>
                  Choose a goal and focus in your learning plan. A short,
                  optional word check helps you choose meanings you already know
                  and words you want to learn. Home suggests a lesson using your
                  current focus and vocabulary.
                </p>
                <p>
                  The current lessons focus on practical A2–B1 English. Your
                  choices guide learning; they are not a test of your English
                  level.
                </p>
                <Link href="/plan" className={link}>
                  Open your learning plan
                </Link>
              </div>
            </details>
            <details>
              <summary className={question}>
                How are lessons and reviews different?
              </summary>
              <div className={answer}>
                <p>
                  Lessons introduce words, check what you remember and practise
                  using them in context. You can leave a lesson and return to
                  its saved step.
                </p>
                <p>
                  Reviews bring back words you have saved. Your answers help
                  schedule when each word returns. If nothing is due, you can
                  learn new words or choose another practice.
                </p>
                <p>
                  Lesson completion appears in Progress. The daily review goal
                  and streak have their own requirements; finishing a lesson
                  does not count as a scheduled review.
                </p>
                <Link href="/progress" className={link}>
                  See your progress
                </Link>
              </div>
            </details>
            <details>
              <summary className={question}>
                What do known, saved and mastered mean?
              </summary>
              <div className={answer}>
                <p>
                  “Already know” is your own assessment of a meaning. Saving a
                  word adds it to your personal vocabulary for review and
                  sentence practice. Mastered is a review stage reached through
                  scheduled practice.
                </p>
                <p>
                  These are separate: marking a meaning as known does not award
                  mastery or add it to your saved words. You can change your
                  known choice and keep private notes on its word page.
                </p>
                <Link href="/vocabulary" className={link}>
                  Explore your vocabulary map
                </Link>
              </div>
            </details>
            <details>
              <summary className={question}>
                Can I practise outside my daily reviews?
              </summary>
              <div className={answer}>
                <p>
                  Yes. Choose typed recall, listening or practice based on
                  mistakes you have made. You can select a lesson’s vocabulary
                  or practise the course mix. These sessions keep their own
                  history.
                </p>
                <p>
                  Listen plays device pronunciation when you ask for it. Check
                  your device’s volume and available English voice if you cannot
                  hear it. Playback does not assess your speaking or
                  pronunciation.
                </p>
                <Link href="/practice" className={link}>
                  Choose a practice
                </Link>
              </div>
            </details>
            <details>
              <summary className={question}>
                How do I set a daily reminder?
              </summary>
              <div className={answer}>
                <p>
                  In Settings, choose a start date and time, then download a
                  calendar reminder and import it into your calendar. Check the
                  first event’s time and alert after importing.
                </p>
                <p>
                  Your calendar controls the reminder and its notifications.
                  Edit or stop it there. Changing the form in Vocanova does not
                  change an imported event, and importing again can create
                  duplicates.
                </p>
                <Link
                  href="/settings#calendar-reminder-heading"
                  className={link}
                >
                  Set up a calendar reminder
                </Link>
              </div>
            </details>
          </div>
        </section>

        <section aria-labelledby="account-help" className="mt-8">
          <h2 id="account-help" className="text-xl font-bold">
            Your writing and account
          </h2>
          <div className="mt-3 divide-y divide-neutral-200 border-y border-neutral-200">
            <details>
              <summary className={question}>
                What happens when I request sentence feedback?
              </summary>
              <div className={answer}>
                <p>
                  Your submitted sentence and the learning context needed to
                  assess it are sent to an AI service. Avoid personal or
                  sensitive details. AI feedback can be wrong; use your own
                  judgment and your teacher’s guidance.
                </p>
                <p>
                  Completed feedback is saved with your sentence history. Draft
                  recovery uses storage in the browser tab when available. You
                  can discard a draft before submitting it.
                </p>
                <p>
                  If a request fails, follow the retry message on the page. You
                  can continue curated lessons and vocabulary reviews while
                  feedback is unavailable. Use “Report feedback” on a result
                  that needs attention.
                </p>
              </div>
            </details>
            <details>
              <summary className={question}>
                Can I download my learning data or close my account?
              </summary>
              <div className={answer}>
                <p>
                  Account security includes a personal data download and account
                  deletion. Download anything you want to keep before closing
                  your account. The deletion screen explains deactivation and
                  the scheduled removal process before you confirm.
                </p>
                <Link href="/settings/account" className={link}>
                  Open account security
                </Link>
              </div>
            </details>
            <details>
              <summary className={question}>Why can’t I sign in?</summary>
              <div className={answer}>
                <p>
                  Use one of the sign-in methods shown on the sign-in page.
                  Access may be limited to invited accounts while Vocanova is
                  being tested. If your session expires, sign in again to
                  continue.
                </p>
                <p>
                  If you still cannot get in, contact support with the message
                  you see. Do not send your password, sign-in link or security
                  code.
                </p>
                <Link href="/login" className={link}>
                  Go to sign in
                </Link>
              </div>
            </details>
          </div>
        </section>

        <section
          aria-labelledby="contact-support"
          className="mt-8 rounded-xl border border-primary-200 bg-primary-50 p-5"
        >
          <h2 id="contact-support" className="text-xl font-bold">
            Contact support
          </h2>
          <p className="mt-2 leading-7 text-neutral-700">
            Tell us what you were trying to do and what happened. Include the
            page name and error message, without private learning content or
            account security details.
          </p>
          <a
            href="mailto:mr.groom.verge@gmail.com"
            className={`${link} mt-2 break-all`}
          >
            mr.groom.verge@gmail.com
          </a>
        </section>
      </div>
    </main>
  );
}
