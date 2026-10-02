import type { WordDetail } from "@vocanova/api-client";

import { formatNoteType, getAdditionalDefinition } from "@/lib/word-definition";
import { ListenButton } from "@/ui/pronunciation";
import { Eyebrow } from "@/ui/surface";

import { formatLevelBand } from "../discover/_components/level-band";
import { MeaningSaveButton } from "../discover/[situation]/[word]/_components/meaning-save-button";
import { formatWordReviewState } from "../discover/[situation]/[word]/_components/word-review-state";
import { SentenceFeedback } from "./sentence-feedback";
import { MeaningKnowledgeEditor } from "./meaning-knowledge-editor";

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
        {contextTitle ? <Eyebrow>{contextTitle}</Eyebrow> : null}
        <h1 className="text-2xl font-semibold text-neutral-900">
          {wordData.text}
        </h1>
        <ListenButton text={wordData.text} />
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          {wordData.wordType.replaceAll("_", " ")}
          {wordData.difficultyLevel
            ? ` · ${formatLevelBand(wordData.difficultyLevel)}`
            : null}
        </p>
      </div>

      <section className="mt-[var(--spacing-lg)]">
        <h2 className="text-xl font-semibold text-neutral-900">Meanings</h2>
        <ul className="mt-[var(--spacing-sm)] space-y-[var(--spacing-md)]">
          {wordData.meanings.map((meaning) => {
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
                className="rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-[var(--spacing-md)]">
                  <div className="min-w-0 flex-1">
                    <p className="font-medium text-neutral-900">
                      {meaning.partOfSpeech}
                    </p>
                    <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
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
                  </div>
                  <MeaningSaveButton
                    meaningId={meaning.id}
                    source={source}
                    initialSaved={meaning.saved}
                    wordText={wordData.text}
                    shortDefinition={meaning.shortDefinition}
                  />
                </div>

                {meaning.examples.length > 0 ? (
                  <div className="mt-[var(--spacing-md)]">
                    <h3 className="text-base font-semibold text-neutral-900">
                      Example sentences
                    </h3>
                    <ul className="mt-[var(--spacing-xs)] list-disc space-y-[var(--spacing-xs)] pl-[var(--spacing-lg)] text-base text-neutral-700">
                      {meaning.examples.map((example) => (
                        <li key={example.id}>
                          <p>{example.exampleText}</p>
                          <ListenButton
                            text={example.exampleText}
                            label={`example: ${example.exampleText}`}
                            showCaption={false}
                          />
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : null}

                {meaning.usageNotes.length > 0 ? (
                  <div className="mt-[var(--spacing-md)]">
                    <h3 className="text-base font-semibold text-neutral-900">
                      Usage notes
                    </h3>
                    <ul className="mt-[var(--spacing-xs)] space-y-[var(--spacing-sm)]">
                      {meaning.usageNotes.map((note) => (
                        <li key={note.id}>
                          <h4 className="text-sm font-semibold text-neutral-800">
                            {formatNoteType(note.noteType)}
                          </h4>
                          <p className="text-base text-neutral-700">
                            {note.noteText}
                          </p>
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : null}

                <MeaningKnowledgeEditor
                  meaningId={meaning.id}
                  initialKnown={meaning.selfReportedKnown}
                />

                {meaning.saved && meaning.userWordId ? (
                  <SentenceFeedback
                    targetWord={wordData.text}
                    attemptId={meaning.userWordId}
                    source="word_detail"
                    userId={userId}
                    shortDefinition={meaning.shortDefinition}
                  />
                ) : null}
              </li>
            );
          })}
        </ul>
      </section>
    </>
  );
}
