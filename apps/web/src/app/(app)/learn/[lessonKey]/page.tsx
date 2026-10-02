import { notFound } from "next/navigation";
import type { LessonSession } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { LessonPlayer } from "../_components/lesson-player";

export default async function LessonPage({
  params,
}: {
  params: Promise<{ lessonKey: string }>;
}) {
  const { lessonKey } = await params;
  const client = await createServerApiClient();
  try {
    const { data } = await client.listLessons();
    const lesson = data.items.find((item) => item.key === lessonKey);
    if (!lesson) notFound();
    let session: LessonSession | null = null;
    if (lesson.sessionId)
      session = (await client.getLessonSession(lesson.sessionId)).data;
    return (
      <PageContainer className="max-w-[44rem]">
        <LessonPlayer lesson={lesson} initialSession={session} />
      </PageContainer>
    );
  } catch (error) {
    requireAuthRedirect(error, `/learn/${encodeURIComponent(lessonKey)}`);
  }
}
