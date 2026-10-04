import type { Metadata } from "next";
import Link from "next/link";
import { ApiResponseError, type WordListsResponse } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer } from "@/ui/surface";
import { ListsCollection } from "./_components/lists-collection";
import { listLink } from "./_components/list-styles";

export const metadata: Metadata = { title: "Personal lists — Vocanova" };
export default async function ListsPage() {
  const client = await createServerApiClient();
  let data: WordListsResponse | null = null;
  try {
    data = (await client.listWordLists()).data;
  } catch (cause) {
    if (cause instanceof ApiResponseError && cause.status === 401)
      requireAuthRedirect(cause, "/lists");
  }
  return (
    <PageContainer>
      <Link href="/discover" className={listLink}>
        Back to Journey
      </Link>
      <header className="my-4">
        <Eyebrow>Words for your life</Eyebrow>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-neutral-900">
          Personal lists
        </h1>
        <p className="mt-2 text-neutral-700">
          Group useful meanings for a trip, a conversation or a project. Lists
          organize your words; saving for review is a separate choice.
        </p>
      </header>
      <ListsCollection initialData={data} />
    </PageContainer>
  );
}
