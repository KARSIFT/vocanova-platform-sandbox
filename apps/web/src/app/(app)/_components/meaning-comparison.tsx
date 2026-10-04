import Link from "next/link";
import type { WordDetail } from "@vocanova/api-client";

export function meaningAnchor(meaningId: string) {
  return `meaning-${meaningId}`;
}

/** Links are derived from real meanings of this word, never guessed relations. */
export function MeaningComparison({
  word,
  currentMeaningId,
  canonicalPath = "",
}: {
  word: WordDetail;
  currentMeaningId?: string;
  canonicalPath?: string;
}) {
  const meanings = word.meanings.filter(
    (meaning) => meaning.id !== currentMeaningId,
  );
  if (word.meanings.length < 2 || !meanings.length) return null;
  return (
    <nav
      aria-label={`Compare meanings of ${word.text}`}
      className="my-5 rounded-xl border border-secondary-200 bg-secondary-50 p-4"
    >
      <h2 className="text-lg font-semibold text-neutral-900">
        {currentMeaningId ? "Other meanings of this word" : "Compare meanings"}
      </h2>
      <p className="mt-2 text-sm text-neutral-700">
        {currentMeaningId
          ? "You saved one meaning. Explore the other meanings separately."
          : "Choose the meaning that fits your situation. Each meaning has its own examples, lists and review choice."}
      </p>
      <ul className="mt-3 space-y-2">
        {meanings.map((meaning, index) => (
          <li key={meaning.id}>
            <Link
              href={`${canonicalPath}#${meaningAnchor(meaning.id)}`}
              className="flex min-h-12 flex-col rounded-lg border border-secondary-200 bg-white px-3 py-2 text-neutral-900 hover:border-primary-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              <span className="text-sm font-semibold text-primary-700">
                {currentMeaningId ? "Explore meaning" : `Meaning ${index + 1}`}{" "}
                · {meaning.partOfSpeech}
              </span>
              <span className="mt-1">{meaning.shortDefinition}</span>
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}
