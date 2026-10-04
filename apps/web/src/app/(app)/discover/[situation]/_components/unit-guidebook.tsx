import Link from "next/link";
import type {
  LessonSummary,
  SituationMeaning,
  StoryCatalogItem,
} from "@vocanova/api-client";
import { Surface } from "@/ui/surface";
import { StoryAudio } from "../../../stories/_components/story-audio";
import { textLink } from "../../../practice/_components/practice-styles";
import { storySituationSlug, unitGuides } from "./unit-guide-content";

export function UnitGuidebook({
  situationSlug,
  meanings,
  examples,
  lessons,
  stories,
}: {
  situationSlug: string;
  meanings: SituationMeaning[];
  examples: Readonly<Record<string, string>>;
  lessons: LessonSummary[] | null;
  stories: StoryCatalogItem[] | null;
}) {
  const guide = unitGuides[situationSlug];
  if (!guide) return null;
  const matchedLessons = lessons?.filter(
    (lesson) => lesson.situationSlug === situationSlug,
  );
  const nextLesson =
    matchedLessons?.find((lesson) => lesson.status === "in_progress") ??
    matchedLessons?.find((lesson) => lesson.status === "not_started") ??
    matchedLessons?.[0];
  const matchedStories =
    stories?.filter(
      (story) => storySituationSlug(story.situation) === situationSlug,
    ) ?? [];
  const words = meanings.filter((meaning) =>
    guide.phrases.some((phrase) => phrase.word === meaning.wordText),
  );
  return (
    <Surface
      id="unit-guide"
      aria-labelledby="unit-guide-heading"
      className="mt-5"
    >
      <h2
        id="unit-guide-heading"
        className="text-xl font-bold text-neutral-900"
      >
        A quick guide
      </h2>
      <p className="mt-2 text-neutral-700">{guide.goal}</p>
      <p className="mt-3 text-sm text-neutral-700">{guide.tip}</p>
      <details className="mt-4 rounded-xl border border-neutral-200 p-4">
        <summary className="min-h-11 cursor-pointer font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
          Useful phrases
        </summary>
        <ol className="mt-3 space-y-4">
          {guide.phrases.map((phrase, index) => (
            <li key={phrase.text}>
              <p className="font-semibold text-neutral-900">{phrase.text}</p>
              <p className="mt-1 text-sm text-neutral-700">{phrase.purpose}</p>
              <StoryAudio
                text={phrase.text}
                label={`guide phrase ${index + 1}`}
              />
            </li>
          ))}
        </ol>
      </details>
      {words.length > 0 && (
        <details className="mt-3 rounded-xl border border-neutral-200 p-4">
          <summary className="min-h-11 cursor-pointer font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
            Words and examples
          </summary>
          <ul className="mt-2 space-y-3">
            {words.map((word) => (
              <li key={word.meaningId}>
                <Link
                  href={`/discover/${encodeURIComponent(situationSlug)}/${encodeURIComponent(word.wordSlug)}#meaning-${encodeURIComponent(word.meaningId)}`}
                  className={textLink}
                >
                  {word.wordText}
                </Link>
                <p className="text-neutral-700">{word.shortDefinition}</p>
                {examples[word.meaningId] && (
                  <p className="mt-2 border-l-2 border-secondary-300 pl-3 text-sm text-neutral-800">
                    {examples[word.meaningId]}
                  </p>
                )}
              </li>
            ))}
          </ul>
        </details>
      )}
      <div className="mt-4 flex flex-wrap gap-x-5 gap-y-2">
        {nextLesson && (
          <Link
            href={`/learn/${encodeURIComponent(nextLesson.key)}`}
            className={textLink}
          >
            {nextLesson.status === "in_progress"
              ? "Continue guided lesson"
              : nextLesson.status === "completed"
                ? "Revisit guided lesson"
                : "Start guided lesson"}
          </Link>
        )}
        {matchedStories.map((story) => (
          <Link
            key={story.key}
            href={`/stories/${encodeURIComponent(story.key)}`}
            className={textLink}
          >
            Story: {story.title}
          </Link>
        ))}
        <Link
          href={`/writing?situation=${encodeURIComponent(situationSlug)}`}
          className={textLink}
        >
          Write about this situation
        </Link>
      </div>
      {lessons === null && (
        <p role="status" className="mt-2 text-sm text-neutral-600">
          Guided lessons could not load. You can still explore this guide and
          its words.
        </p>
      )}
      {stories === null && (
        <p role="status" className="mt-2 text-sm text-neutral-600">
          Story links could not load. You can try the story library later.
        </p>
      )}
    </Surface>
  );
}
