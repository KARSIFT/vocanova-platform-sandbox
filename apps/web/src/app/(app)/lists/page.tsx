import type { Metadata } from "next";
import { ApiResponseError, type WordListsResponse } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader } from "@/ui/surface";
import { ListsCollection } from "./_components/lists-collection";

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
      <PageHeader
        title="Personal lists"
        back={{ href: "/discover", label: "Back to Journey" }}
        lede="Group useful meanings for a trip, a conversation or a project."
      />
      <ListsCollection initialData={data} />
    </PageContainer>
  );
}
