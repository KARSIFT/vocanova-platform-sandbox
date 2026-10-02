import Link from "next/link";
import type { VocabularySearchItem } from "@vocanova/api-client";

function stage(item: VocabularySearchItem) {
  if (item.selfReportedKnown)
    return {
      label: "Already known",
      color: "border-primary-300 bg-primary-50",
      symbol: "✓",
    };
  if (item.saved)
    return {
      label: "Saved to learn",
      color: "border-secondary-300 bg-secondary-50",
      symbol: "+",
    };
  return {
    label: "Not explored",
    color: "border-neutral-200 bg-white",
    symbol: "○",
  };
}

export function VocabularyMap({ items }: { items: VocabularySearchItem[] }) {
  return (
    <section aria-label="Word knowledge map">
      <p className="mb-4 text-sm text-neutral-600">
        Each tile is one meaning in our collection. Open it to explore, save it
        to learn, or mark it as already known. Knowing a word is your
        assessment, separate from its review stage.
      </p>
      <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {items.map((item) => {
          const status = stage(item);
          return (
            <li key={item.meaningId}>
              <Link
                href={`/vocabulary/${encodeURIComponent(item.wordSlug)}`}
                className={`flex h-full min-h-40 flex-col rounded-2xl border p-4 text-neutral-900 transition-colors hover:border-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 ${status.color}`}
              >
                <span
                  aria-hidden="true"
                  className="text-xl font-bold text-primary-700"
                >
                  {status.symbol}
                </span>
                <span className="mt-2 break-words text-lg font-bold">
                  {item.wordText}
                </span>
                <span className="text-xs text-neutral-600">
                  {item.partOfSpeech}
                </span>
                <span className="mt-2 grow text-sm">
                  {item.shortDefinition}
                </span>
                <span className="mt-4 text-sm font-semibold">
                  {status.label}
                </span>
                {item.selfReportedKnown && item.saved && (
                  <span className="mt-1 text-xs text-neutral-600">
                    Also saved for practice
                  </span>
                )}
                {item.due && (
                  <span className="mt-1 text-xs text-neutral-600">
                    Due for review
                  </span>
                )}
              </Link>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
