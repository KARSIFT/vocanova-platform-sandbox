import type { Metadata } from "next";
import Link from "@/ui/link";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader, PageSection, Surface } from "@/ui/surface";
import { textLink } from "../../practice/_components/practice-styles";
import { StoryAudio } from "../_components/story-audio";
import { StoryStart } from "../_components/story-start";
import { StoryVocabularyHelp } from "../_components/story-vocabulary";
import { storySituationSlug } from "../../discover/[situation]/_components/unit-guide-content";

import styles from "../stories.module.css";

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
    <PageContainer>
      <PageHeader
        back={{ href: "/stories", label: "Back to Stories" }}
        eyebrow={`${story.situation} · ${story.level}`}
        title={story.title}
        lede={story.description}
      />
      <Surface aria-label="Story practice">
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
      </Surface>
      <StoryVocabularyHelp vocabulary={story.vocabulary} />
      <PageSection
        className="mt-8"
        title="Full story"
        titleId="full-story-heading"
        description="Audio is optional. You can read every line."
      >
        <ol className={styles.dialogue}>
          {story.lines.map((line, index) => (
            <li key={line.id} className={styles.dialogueLine}>
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
      </PageSection>
    </PageContainer>
  );
}
