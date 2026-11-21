import type { Metadata } from "next";
import Link from "@/ui/link";

import { BrandMark } from "@/ui/brand-mark";
import { PageHeader, PageSection } from "@/ui/surface";
import styles from "../public-page.module.css";

export const metadata: Metadata = {
  title: "Help — Vocanova",
  description: "Find your next learning step and get help with Vocanova.",
};

export default function HelpPage() {
  return (
    <main className={styles.canvas}>
      <div className={`${styles.container} ${styles.helpContainer}`}>
        <header className={styles.header}>
          <Link href="/" aria-label="VocaNova home" className={styles.brand}>
            <BrandMark />
          </Link>
          <Link href="/home" className={styles.textLink}>
            Go to learning
          </Link>
        </header>
        <PageHeader
          title="How can we help?"
          lede="Find a useful next step, understand your progress, or get in touch."
          className={styles.helpIntro}
        />

        <div className={styles.helpContent}>
          <div className="min-w-0">
            <PageSection title="Learning with Vocanova" titleId="learning-help">
              <div className={styles.helpQuestions}>
                <details open>
                  <summary>Where should I start?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Choose a goal and focus in your learning plan. A short,
                      optional word check helps you choose meanings you already
                      know and words you want to learn. Home helps you review
                      saved words, discover new ones, and choose a short
                      practice lesson.
                    </p>
                    <p>
                      The current lessons focus on practical A2–B1 English. Your
                      choices guide learning; they are not a test of your
                      English level.
                    </p>
                    <Link href="/plan" className={styles.textLink}>
                      Open your learning plan
                    </Link>
                  </div>
                </details>
                <details>
                  <summary>How are lessons and reviews different?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Lessons introduce words, check what you remember and
                      practise using them in context. You can leave a lesson and
                      return to its saved step.
                    </p>
                    <p>
                      Reviews bring back words you have saved. Your answers help
                      schedule when each word returns. If nothing is due, you
                      can learn new words or choose another practice.
                    </p>
                    <p>
                      Lesson completion appears in Progress. The daily review
                      goal and streak have their own requirements; finishing a
                      lesson does not count as a scheduled review.
                    </p>
                    <Link href="/progress" className={styles.textLink}>
                      See your progress
                    </Link>
                  </div>
                </details>
                <details>
                  <summary>What do known, saved and mastered mean?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      “Already know” is your own assessment of a meaning. Saving
                      a word adds it to your personal vocabulary for review and
                      sentence practice. Mastered is a review stage reached
                      through scheduled practice.
                    </p>
                    <p>
                      These are separate: marking a meaning as known does not
                      award mastery or add it to your saved words. You can
                      change your known choice and keep private notes on its
                      word page.
                    </p>
                    <Link href="/discover" className={styles.textLink}>
                      Explore your word map
                    </Link>
                  </div>
                </details>
                <details>
                  <summary>Can I practise outside my daily reviews?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Yes. Choose typed recall, listening or practice based on
                      mistakes you have made. You can select a lesson’s
                      vocabulary or practise the course mix. These sessions keep
                      their own history.
                    </p>
                    <p>
                      Listen plays device pronunciation when you ask for it.
                      Check your device’s volume and available English voice if
                      you cannot hear it. Playback does not assess your speaking
                      or pronunciation.
                    </p>
                    <Link href="/practice" className={styles.textLink}>
                      Choose a practice
                    </Link>
                  </div>
                </details>
                <details>
                  <summary>How do I set a daily reminder?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      In Settings, choose a start date and time, then download a
                      calendar reminder and import it into your calendar. Check
                      the first event’s time and alert after importing.
                    </p>
                    <p>
                      Your calendar controls the reminder and its notifications.
                      Edit or stop it there. Changing the form in Vocanova does
                      not change an imported event, and importing again can
                      create duplicates.
                    </p>
                    <Link
                      href="/settings#calendar-reminder-heading"
                      className={styles.textLink}
                    >
                      Set up a calendar reminder
                    </Link>
                  </div>
                </details>
              </div>
            </PageSection>
            <PageSection
              title="Your writing and account"
              titleId="account-help"
            >
              <div className={styles.helpQuestions}>
                <details>
                  <summary>
                    What happens when I request sentence feedback?
                  </summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Your submitted sentence and the learning context needed to
                      assess it are sent to an AI service. Avoid personal or
                      sensitive details. AI feedback can be wrong; use your own
                      judgment and your teacher’s guidance.
                    </p>
                    <p>
                      Completed feedback is saved with your sentence history.
                      Draft recovery uses storage in the browser tab when
                      available. You can discard a draft before submitting it.
                    </p>
                    <p>
                      If a request fails, follow the retry message on the page.
                      You can continue curated lessons and vocabulary reviews
                      while feedback is unavailable. Use “Report feedback” on a
                      result that needs attention.
                    </p>
                  </div>
                </details>
                <details>
                  <summary>
                    Can I download my learning data or close my account?
                  </summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Account security includes a personal data download and
                      account deletion. Download anything you want to keep
                      before closing your account. The deletion screen explains
                      deactivation and the scheduled removal process before you
                      confirm.
                    </p>
                    <Link href="/settings/account" className={styles.textLink}>
                      Open account security
                    </Link>
                  </div>
                </details>
                <details>
                  <summary>Why can’t I sign in?</summary>
                  <div className={styles.helpAnswer}>
                    <p>
                      Use one of the sign-in methods shown on the sign-in page.
                      Access may be limited to invited accounts while Vocanova
                      is being tested. If your session expires, sign in again to
                      continue.
                    </p>
                    <p>
                      If you still cannot get in, contact support with the
                      message you see. Do not send your password, sign-in link
                      or security code.
                    </p>
                    <Link href="/login" className={styles.textLink}>
                      Go to sign in
                    </Link>
                  </div>
                </details>
              </div>
            </PageSection>
          </div>
          <PageSection
            title="Contact support"
            titleId="contact-support"
            className={styles.helpContact}
          >
            <p>
              Tell us what you were trying to do and what happened. Include the
              page name and error message, without private learning content or
              account security details.
            </p>
            <a
              href="mailto:mr.groom.verge@gmail.com"
              className={styles.textLink}
            >
              mr.groom.verge@gmail.com
            </a>
          </PageSection>
        </div>
      </div>
    </main>
  );
}
