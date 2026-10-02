import Link from "next/link";
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
      <Link
        href="/vocabulary"
        className="inline-flex min-h-11 items-center text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
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
