import Link from "next/link";
import type { KnowledgeSummary } from "@vocanova/api-client";
import { Surface } from "@/ui/surface";

export function KnowledgeOverview({ summary }: { summary: KnowledgeSummary }) {
  const stages = [
    ["New", summary.new, "Saved and ready to start"],
    ["Learning", summary.learning, "Building recall"],
    ["Reviewing", summary.reviewing, "Keeping words fresh"],
    ["Mastered", summary.mastered, "Established through your review history"],
  ] as const;
  return (
    <Surface aria-labelledby="vocabulary-map-heading" className="mt-6">
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
        {stages.map(([label, count, explanation]) => (
          <div key={label} className="border-l-4 border-primary-200 pl-3">
            <dt className="font-semibold text-neutral-700">{label}</dt>
            <dd className="mt-1 text-3xl font-bold tabular-nums text-neutral-900">
              {count}
            </dd>
            <dd className="mt-1 text-sm text-neutral-600">{explanation}</dd>
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
      <p className="mt-3 text-sm text-neutral-600">
        Stages describe your saved vocabulary, not your overall English level.
      </p>
    </Surface>
  );
}
