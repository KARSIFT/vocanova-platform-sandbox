import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import { textLink } from "../../practice/_components/practice-styles";
import { StoryAudio } from "../_components/story-audio";
import { StoryStart } from "../_components/story-start";
import { StoryVocabularyHelp } from "../_components/story-vocabulary";
import { storySituationSlug } from "../../discover/[situation]/_components/unit-guide-content";

export const metadata: Metadata = { title: "Read a story — Vocanova" };
export default async function StoryReadingPage({
  params,
}: {
  params: Promise<{ storyKey: string }>;
}) {
  const { storyKey } = await params;
  let story;
  try {
    story = (await (await createServerApiClient()).getStory(storyKey)).data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) notFound();
    requireAuthRedirect(error, `/stories/${encodeURIComponent(storyKey)}`);
  }
  const situationSlug = storySituationSlug(story.situation);
  return (
    <PageContainer className="max-w-[44rem]">
      <Link href="/stories" className={textLink}>
        Back to Stories
      </Link>
      <Surface className="mt-3">
        <Eyebrow>
          {story.situation} · {story.level}
        </Eyebrow>
        <h1 className="mt-2 text-3xl font-bold text-neutral-900">
          {story.title}
        </h1>
        <p className="my-4 text-neutral-700">{story.description}</p>
        <StoryStart storyKey={story.key} />
        <p className="mt-3 text-sm text-neutral-600">
          Practice reveals the dialogue a line at a time and saves each step.
          You can also read the full story below.
        </p>
        {situationSlug && (
          <Link
            href={`/discover/${encodeURIComponent(situationSlug)}#unit-guide`}
            className={`${textLink} mt-3`}
          >
            Open the situation guide
          </Link>
        )}
        <StoryVocabularyHelp vocabulary={story.vocabulary} />
        <h2 className="mb-3 mt-6 text-xl font-bold text-neutral-900">
          Full story
        </h2>
        <ol className="space-y-4">
          {story.lines.map((line, index) => (
            <li key={line.id} className="rounded-xl bg-neutral-50 p-4">
              <p className="text-sm font-semibold text-primary-700">
                {line.speaker}
              </p>
              <p className="mt-1 text-lg leading-relaxed text-neutral-900">
                {line.text}
              </p>
              <StoryAudio
                text={line.text}
                label={`line ${index + 1} by ${line.speaker}`}
              />
            </li>
          ))}
        </ol>
      </Surface>
    </PageContainer>
  );
}
