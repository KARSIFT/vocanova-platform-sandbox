import type { WordDetail } from "@vocanova/api-client";

import { getAdditionalDefinition } from "@/lib/word-definition";
import { MeaningPicture } from "@/ui/meaning-picture";
import { ListenButton } from "@/ui/pronunciation";
import { WordPracticeDisclosure } from "./word-practice-disclosure";

import { formatLevelBand } from "../discover/_components/level-band";
import { MeaningSaveButton } from "../discover/[situation]/[word]/_components/meaning-save-button";
import { formatWordReviewState } from "../discover/[situation]/[word]/_components/word-review-state";
import { SentenceFeedback } from "./sentence-feedback";
import { MeaningKnowledgeEditor } from "./meaning-knowledge-editor";
import { MeaningListEditor } from "./meaning-list-editor";
import { MeaningTeaching } from "./meaning-teaching";
import { MeaningComparison, meaningAnchor } from "./meaning-comparison";

interface WordDetailContentProps {
  word: WordDetail;
  userId?: string;
  contextTitle?: string;
  source: "journey" | "search";
}

export function WordDetailContent({
  word: wordData,
  userId,
  contextTitle,
  source,
}: WordDetailContentProps) {
  return (
    <>
      <div className="mt-[var(--spacing-md)]">
        {contextTitle ? (
          <p className="mb-3 inline-flex rounded-full bg-secondary-50 px-3 py-1 text-sm font-semibold text-secondary-800">
            {contextTitle}
          </p>
        ) : null}
        {/* Set like a dictionary entry: headword, then its kind of word. */}
        <h1 className="headword text-[clamp(2.6rem,8vw,3.6rem)]">
          {wordData.text}
        </h1>
        {wordData.meanings.length === 0 && (
          <ListenButton text={wordData.text} />
        )}
        <p className="mt-[var(--spacing-sm)] flex flex-wrap items-center gap-x-3 gap-y-1 text-base">
          <span className="pos">{wordData.wordType.replaceAll("_", " ")}</span>
          {wordData.difficultyLevel ? (
            <span className="rounded-full bg-secondary-50 px-2.5 py-0.5 text-sm font-semibold text-secondary-800">
              {formatLevelBand(wordData.difficultyLevel)}
            </span>
          ) : null}
        </p>
      </div>

      <MeaningComparison word={wordData} />

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-bold text-neutral-900">Meanings</h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-md)]">
          {wordData.meanings.map((meaning, index) => {
            const additionalDefinition = getAdditionalDefinition(
              meaning.shortDefinition,
              meaning.learnerDefinition,
            );
            const reviewState = formatWordReviewState(
              meaning.reviewState,
              meaning.due,
            );

            return (
              <li
                key={meaning.id}
                id={meaningAnchor(meaning.id)}
                tabIndex={-1}
                className="scroll-mt-24 rounded-[1.5rem] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-[var(--vn-shadow-card)] sm:p-[var(--spacing-lg)]"
              >
                {/* Kind of word and the save toggle share one row so the
                    definition below keeps the full reading width. */}
                <div className="flex items-start justify-between gap-[var(--spacing-md)]">
                  <p className="pos pt-2.5 text-base">{meaning.partOfSpeech}</p>
                  <MeaningSaveButton
                    meaningId={meaning.id}
                    source={source}
                    initialSaved={meaning.saved}
                    wordText={wordData.text}
                    shortDefinition={meaning.shortDefinition}
                  />
                </div>
                <p className="mt-[var(--spacing-xs)] text-xl leading-snug text-neutral-900">
                  {meaning.shortDefinition}
                </p>
                {additionalDefinition ? (
                  <p className="mt-[var(--spacing-xs)] text-base text-neutral-600">
                    {additionalDefinition}
                  </p>
                ) : null}
                {reviewState ? (
                  <p className="mt-[var(--spacing-xs)] text-sm font-medium text-primary-700">
                    {reviewState}
                  </p>
                ) : null}
                <ListenButton text={wordData.text} meaningId={meaning.id} />

                <MeaningPicture
                  meaningId={meaning.id}
                  image={meaning.image}
                  meaningRevision={meaning.meaningRevision}
                />
                <MeaningTeaching meaning={meaning} />

                <details className="mt-4 border-t border-neutral-200 pt-2">
                  <summary className="min-h-12 cursor-pointer content-center font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
                    Personal tools
                  </summary>
                  <MeaningListEditor meaningId={meaning.id} />
                  <MeaningKnowledgeEditor
                    meaningId={meaning.id}
                    initialKnown={meaning.selfReportedKnown}
                  />
                  {meaning.saved && meaning.userWordId ? (
                    <WordPracticeDisclosure
                      anchor={
                        index === 0
                          ? "sentence-practice"
                          : `sentence-practice-${meaning.id}`
                      }
                    >
                      <SentenceFeedback
                        targetWord={wordData.text}
                        attemptId={meaning.userWordId}
                        source="word_detail"
                        userId={userId}
                        shortDefinition={meaning.shortDefinition}
                      />
                    </WordPracticeDisclosure>
                  ) : null}
                </details>
              </li>
            );
          })}
        </ul>
      </section>
    </>
  );
}
