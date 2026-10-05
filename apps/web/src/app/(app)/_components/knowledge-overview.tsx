import Link from "next/link";
import type { KnowledgeSummary } from "@vocanova/api-client";

export function KnowledgeOverview({ summary }: { summary: KnowledgeSummary }) {
  const stages = [
    ["New", summary.new, "Saved and ready to start"],
    ["Learning", summary.learning, "Building recall"],
    ["Reviewing", summary.reviewing, "Keeping words fresh"],
    ["Mastered", summary.mastered, "Established through your review history"],
  ] as const;
  return (
    <section
      aria-labelledby="vocabulary-map-heading"
      className="mt-6 rounded-2xl bg-white p-5 sm:p-7"
    >
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id="vocabulary-map-heading"
            className="text-2xl font-bold text-neutral-900"
          >
            Your vocabulary map
          </h2>
          <p className="mt-2 text-neutral-700">
            {summary.saved} saved {summary.saved === 1 ? "meaning" : "meanings"}
            , each at its own stage.
          </p>
        </div>
        <Link
          href="/vocabulary?view=map"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700"
        >
          Explore knowledge map
        </Link>
      </div>
      <dl className="mt-5 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stages.map(([label, count]) => (
          <div key={label} className="border-l-4 border-primary-200 pl-3">
            <dt className="font-semibold text-neutral-700">{label}</dt>
            <dd className="mt-1 text-3xl font-bold tabular-nums text-neutral-900">
              {count}
            </dd>
          </div>
        ))}
      </dl>
      <p className="mt-5 rounded-xl bg-primary-50 p-4 text-neutral-900">
        <Link
          href="/vocabulary?view=map&knowledge=known"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 underline underline-offset-4"
        >
          {summary.selfReportedKnown}{" "}
          {summary.selfReportedKnown === 1 ? "meaning" : "meanings"} marked
          already known
        </Link>
        <span className="mt-1 block text-sm">
          Your own assessment, counted separately. These meanings may also be
          saved for practice.
        </span>
      </p>
      {(summary.ignored > 0 || summary.archived > 0) && (
        <p className="mt-4 text-sm text-neutral-600">
          Also saved: {summary.ignored} paused and {summary.archived} archived
          meanings.
        </p>
      )}
      <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-neutral-200 pt-4">
        <p className="text-neutral-700">
          {summary.due
            ? `${summary.due} ready for review now`
            : "You’re up to date with scheduled reviews."}
        </p>
        <Link
          href={summary.due ? "/review" : "/vocabulary"}
          className="inline-flex min-h-11 items-center font-semibold text-primary-700"
        >
          {summary.due ? "Start review" : "Find new words"}
        </Link>
      </div>
      <details className="mt-3 border-t border-neutral-200 pt-2">
        <summary className="min-h-11 cursor-pointer content-center text-sm font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
          About these stages
        </summary>
        <dl className="space-y-2 pb-3 text-sm text-neutral-600">
          {stages.map(([label, , explanation]) => (
            <div key={label}>
              <dt className="font-semibold text-neutral-900">{label}</dt>
              <dd>{explanation}</dd>
            </div>
          ))}
        </dl>
        <p className="pb-3 text-sm text-neutral-600">
          Stages describe your saved vocabulary, not your overall English level.
        </p>
      </details>
    </section>
  );
}
