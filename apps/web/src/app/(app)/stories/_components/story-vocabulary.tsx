import Link from "@/ui/link";
import type { StoryVocabulary } from "@vocanova/api-client";
import { textLink } from "../../practice/_components/practice-styles";

export function StoryVocabularyHelp({
  vocabulary,
}: {
  vocabulary: StoryVocabulary[];
}) {
  return (
    <details className="mt-5 rounded-xl border border-neutral-200 bg-white px-4 open:pb-4">
      <summary className="vn-summary py-3">Vocabulary help</summary>
      <ul className="divide-y divide-neutral-200">
        {vocabulary.map((word) => (
          <li key={word.meaningId} className="pb-3">
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
