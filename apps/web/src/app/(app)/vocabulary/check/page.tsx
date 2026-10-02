import type { Metadata } from "next";
import Link from "next/link";
import {
  ApiResponseError,
  type VocabularySearchItem,
} from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { VocabularySelfCheck } from "./_components/self-check";

export const metadata: Metadata = {
  title: "Find your starting words — Vocanova",
};

export default async function VocabularyCheckPage() {
  const client = await createServerApiClient();
  let items: VocabularySearchItem[] | null = null;
  let category: string | undefined;
  try {
    const profile = (await client.getLearningPreferences()).data;
    category = profile.mainUseCase ?? undefined;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/vocabulary/check");
  }
  try {
    let response = await client.searchVocabulary({
      knowledge: "unexplored",
      category,
      limit: 10,
    });
    if (category && response.data.items.length === 0) {
      category = undefined;
      response = await client.searchVocabulary({
        knowledge: "unexplored",
        limit: 10,
      });
    }
    items = response.data.items.slice(0, 10);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/vocabulary/check");
  }
  return (
    <PageContainer className="max-w-[48rem]">
      <Link
        href="/vocabulary"
        className="inline-flex min-h-11 items-center font-semibold text-primary-700"
      >
        Back to vocabulary
      </Link>
      <header className="mb-5 mt-3">
        <h1 className="text-xl font-bold tracking-tight text-neutral-900 sm:text-3xl">
          Find your starting words
        </h1>
        <p className="mt-2 max-w-[40rem] text-neutral-700">
          Choose meanings you know or want to learn. This check is optional.
        </p>
        {category && (
          <p className="mt-2 text-sm text-neutral-600">
            Starting with your focus: {category.replaceAll("_", " ")}.
          </p>
        )}
      </header>
      <VocabularySelfCheck
        key={items?.map((item) => item.meaningId).join(",") ?? "unavailable"}
        items={items}
      />
      <p className="mt-4 text-sm text-neutral-600">
        Saved and already-known meanings are left out. Your private notes stay
        unchanged. You can leave the check at any time.
      </p>
      <Link
        href="/plan"
        className="mt-5 inline-flex min-h-11 items-center font-semibold text-primary-700"
      >
        Go to your learning plan
      </Link>
    </PageContainer>
  );
}
