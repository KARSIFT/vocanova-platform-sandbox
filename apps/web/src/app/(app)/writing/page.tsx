import type { Metadata } from "next";
import Link from "@/ui/link";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { getPrimaryDefinition } from "@/lib/word-definition";
import { PageContainer, PageHeader, Surface } from "@/ui/surface";
import { SaveAndWrite } from "./_components/save-and-write";
import { writingPrompts } from "./_components/writing-prompts";
import { textLink } from "../practice/_components/practice-styles";

export const metadata: Metadata = {
  title: "Topic writing — Vocanova",
  description: "Use useful English words in original everyday writing prompts.",
};
export default async function WritingPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const situation = typeof query.situation === "string" ? query.situation : "";
  const meaningId = typeof query.meaning === "string" ? query.meaning : "";
  const client = await createServerApiClient();
  const returnTo = `/writing${situation ? `?${new URLSearchParams({ situation, ...(meaningId ? { meaning: meaningId } : {}) })}` : ""}`;
  let situations;
  let detail;
  let currentUser;
  try {
    const [library, user, situationDetail] = await Promise.all([
      client.listJourneySituations({ limit: 50 }),
      client.getCurrentUser(),
      situation ? client.getJourneySituation(situation) : undefined,
    ]);
    situations = library.data;
    currentUser = user.data;
    detail = situationDetail?.data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 404) notFound();
    requireAuthRedirect(error, returnTo);
  }
  const selected = detail?.meanings.find(
    (item) => item.meaningId === meaningId,
  );
  if (meaningId && !selected) notFound();
  let canonical;
  if (selected) {
    try {
      canonical = (
        await client.getCanonicalWord(selected.wordSlug)
      ).data.word.meanings.find((item) => item.id === selected.meaningId);
    } catch (error) {
      requireAuthRedirect(error, returnTo);
    }
    if (!canonical) notFound();
  }
  return (
    <PageContainer>
      <PageHeader
        title="Topic writing"
        back={{ href: "/practice", label: "Back to Practice" }}
        lede="Choose a topic and a word, then get feedback on a sentence you could use today."
      />
      {situations.items.length > 0 && (
        <details
          open={!situation}
          className="mb-4 rounded-xl border border-neutral-200 bg-white px-4"
        >
          <summary className="vn-summary py-3">
            {detail
              ? `Change topic: ${detail.situation.title}`
              : "Choose a writing topic"}
          </summary>
          <nav
            aria-label="Writing situations"
            className="flex flex-wrap gap-2 pb-4"
          >
            {situations.items.map((item) => (
              <Link
                key={item.id}
                href={`/writing?${new URLSearchParams({ situation: item.slug })}`}
                aria-current={item.slug === situation ? "page" : undefined}
                className="vn-chip-link max-w-full [overflow-wrap:anywhere]"
              >
                {item.title}
              </Link>
            ))}
          </nav>
        </details>
      )}
      {situations.hasMore && (
        <Link href="/courses" className={textLink}>
          Explore more situations
        </Link>
      )}
      {detail ? (
        <Surface>
          <h2 className="text-2xl font-bold text-neutral-900">
            {detail.situation.title}
          </h2>
          <p className="mt-3 rounded-xl bg-secondary-50 p-4 text-lg leading-relaxed text-neutral-900">
            {writingPrompts[situation] ??
              "Write one sentence about this situation using a word you want to practise."}
          </p>
          <details
            open={!selected}
            className="mt-4 border-y border-neutral-200"
          >
            <summary className="vn-summary scroll-my-24 py-3">
              {selected
                ? `Change meaning: ${selected.wordText}`
                : "Choose a meaning"}
            </summary>
            {detail.meanings.length ? (
              <ul className="divide-y divide-neutral-200">
                {detail.meanings.map((item) => (
                  <li key={item.meaningId} className="pb-3">
                    <Link
                      className={textLink}
                      aria-current={
                        item.meaningId === meaningId ? "page" : undefined
                      }
                      href={`/writing?${new URLSearchParams({ situation, meaning: item.meaningId })}`}
                    >
                      {item.wordText}
                    </Link>
                    <p className="text-sm text-neutral-700">
                      {getPrimaryDefinition(
                        item.shortDefinition,
                        item.learnerDefinition,
                      )}
                    </p>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-3 text-neutral-700">
                No meanings are available here yet. Choose another situation.
              </p>
            )}
          </details>
          {selected && canonical && (
            <section
              aria-label="Writing with your selected meaning"
              className="mt-6"
            >
              <h3
                id={`writing-meaning-heading-${selected.meaningId}`}
                className="section-title"
              >
                Write with {selected.wordText}
              </h3>
              {canonical.shortDefinition ? (
                <p className="mt-2 text-neutral-700">
                  {getPrimaryDefinition(
                    canonical.shortDefinition,
                    canonical.learnerDefinition,
                  )}
                </p>
              ) : null}
              <SaveAndWrite
                key={`${currentUser.id}:${selected.meaningId}`}
                meaningId={selected.meaningId}
                wordSlug={selected.wordSlug}
                wordText={selected.wordText}
                shortDefinition={canonical.shortDefinition}
                introHeadingId={`writing-meaning-heading-${selected.meaningId}`}
                initialMeaning={canonical}
                userId={currentUser.id}
              />
              <p className="mt-4 text-sm text-neutral-600">
                Feedback checks the chosen meaning and language, rather than
                whether you followed the topic. Checked sentences appear in your
                writing history.
              </p>
            </section>
          )}
          <Link
            href={`/discover/${encodeURIComponent(situation)}`}
            className={`${textLink} mt-6`}
          >
            Explore this situation
          </Link>
        </Surface>
      ) : !situations.items.length ? (
        <div className="vn-empty">
          <h2>No writing topics are available</h2>
          <p className="mt-3 text-neutral-700">
            You can still write sentences with your saved words. Try the topic
            library again later.
          </p>
          <Link href="/words" className={textLink}>
            Practise with saved words
          </Link>
        </div>
      ) : null}
      <div className="mt-6 flex flex-wrap gap-4">
        <Link href="/progress/sentences" className={textLink}>
          Your writing history
        </Link>
        <Link href="/words" className={textLink}>
          Your saved vocabulary
        </Link>
      </div>
    </PageContainer>
  );
}
