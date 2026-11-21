import Link from "@/ui/link";
import { notFound } from "next/navigation";

import { ApiResponseError } from "@vocanova/api-client";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { WordDetailContent } from "../../_components/word-detail-content";

export default async function VocabularyWordPage({
  params,
}: {
  params: Promise<{ word: string }>;
}) {
  const { word } = await params;
  const client = await createServerApiClient();
  let wordResponse: Awaited<ReturnType<typeof client.getCanonicalWord>>;
  let userResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [wordResponse, userResponse] = await Promise.all([
      client.getCanonicalWord(word),
      client.getCurrentUser(),
    ]);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) {
      notFound();
    }
    requireAuthRedirect(error, `/vocabulary/${word}`);
  }

  return (
    <PageContainer>
      <Link href="/vocabulary" className="page-back">
        <svg aria-hidden="true" viewBox="0 0 24 24">
          <path
            d="m15 5-7 7 7 7"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        Back to word search
      </Link>
      <WordDetailContent
        word={wordResponse.data.word}
        userId={userResponse.data.id}
        source="search"
      />
    </PageContainer>
  );
}
