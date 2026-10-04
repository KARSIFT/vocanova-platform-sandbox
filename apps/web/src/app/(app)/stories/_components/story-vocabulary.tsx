import Link from "next/link";
import type { StoryVocabulary } from "@vocanova/api-client";
import { textLink } from "../../practice/_components/practice-styles";

export function StoryVocabularyHelp({
  vocabulary,
}: {
  vocabulary: StoryVocabulary[];
}) {
  return (
    <details className="mt-5 rounded-xl border border-neutral-200 p-4">
      <summary className="min-h-11 cursor-pointer font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
        Vocabulary help
      </summary>
      <ul className="mt-3 space-y-3">
        {vocabulary.map((word) => (
          <li key={word.meaningId}>
            <Link
              href={`/vocabulary/${encodeURIComponent(word.wordSlug)}#meaning-${encodeURIComponent(word.meaningId)}`}
              className={textLink}
            >
              {word.wordText}
            </Link>
            <p className="text-neutral-700">{word.definition}</p>
          </li>
        ))}
      </ul>
      <p className="mt-3 text-sm text-neutral-600">
        Read the meaning or open its word page. Saving words for review is your
        choice.
      </p>
    </details>
  );
}
