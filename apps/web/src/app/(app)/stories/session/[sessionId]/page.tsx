import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { StoryPlayer } from "../../_components/story-player";

export const metadata: Metadata = { title: "Story practice — Vocanova" };
export default async function StorySessionPage({
  params,
}: {
  params: Promise<{ sessionId: string }>;
}) {
  const { sessionId } = await params;
  try {
    const { data } = await (
      await createServerApiClient()
    ).getStorySession(sessionId);
    return (
      <PageContainer className="max-w-[44rem]">
        <StoryPlayer key={data.id} initialSession={data} />
      </PageContainer>
    );
  } catch (error) {
    if (
      error instanceof ApiResponseError &&
      (error.status === 404 || error.status === 422)
    )
      notFound();
    requireAuthRedirect(
      error,
      `/stories/session/${encodeURIComponent(sessionId)}`,
    );
  }
}
