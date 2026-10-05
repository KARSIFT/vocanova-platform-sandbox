import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import { SaveAndWrite } from "./_components/save-and-write";
import { writingPrompts } from "./_components/writing-prompts";
import {
  textLink,
  secondaryAction,
} from "../practice/_components/practice-styles";

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
    const [library, user] = await Promise.all([
      client.listJourneySituations({ limit: 50 }),
      client.getCurrentUser(),
    ]);
    situations = library.data;
    currentUser = user.data;
    if (situation) detail = (await client.getJourneySituation(situation)).data;
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
    <PageContainer className="max-w-[64rem]">
      <Link href="/practice" className={textLink}>
        Back to Practice
      </Link>
      <header className="my-6">
        <Eyebrow>Use your words</Eyebrow>
        <h1 className="mt-2 text-3xl font-bold text-neutral-900">
          Topic writing
        </h1>
        <p className="mt-3 max-w-[44rem] text-lg text-neutral-700">
          Write something you could use today. Choose a topic and a word, then
          compare feedback with your own sentence.
        </p>
      </header>
      <details
        open={!situation}
        className="mb-6 rounded-xl border border-neutral-200 bg-white p-4"
      >
        <summary className="min-h-11 cursor-pointer py-3 font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
          {detail
            ? `Change topic: ${detail.situation.title}`
            : "Choose a writing topic"}
        </summary>
        <nav
          aria-label="Writing situations"
          className="mb-6 flex flex-wrap gap-3"
        >
          {situations.items.map((item) => (
            <Link
              key={item.id}
              href={`/writing?${new URLSearchParams({ situation: item.slug })}`}
              aria-current={item.slug === situation ? "page" : undefined}
              className={`${secondaryAction} ${item.slug === situation ? "border-primary-600 bg-primary-50" : ""}`}
            >
              {item.title}
            </Link>
          ))}
        </nav>
      </details>
      {situations.hasMore && (
        <Link href="/discover" className={textLink}>
          Explore more situations
        </Link>
      )}
      {detail ? (
        <Surface>
          <h2 className="text-2xl font-bold text-neutral-900">
            {detail.situation.title}
          </h2>
          <p className="mt-3 text-lg text-neutral-800">
            {writingPrompts[situation] ??
              "Write one sentence about this situation using a word you want to practise."}
          </p>
          <details
            open={!selected}
            className="mt-5 rounded-xl border border-neutral-200 p-4"
          >
            <summary className="min-h-11 cursor-pointer py-3 font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
              {selected
                ? `Change meaning: ${selected.wordText}`
                : "Choose a meaning"}
            </summary>
            {detail.meanings.length ? (
              <ul className="my-3 grid gap-3 sm:grid-cols-2">
                {detail.meanings.map((item) => (
                  <li
                    key={item.meaningId}
                    className="rounded-xl border border-neutral-200 p-3"
                  >
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
                      {item.shortDefinition}
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
              className="mt-6 border-t border-neutral-200 pt-6"
            >
              <h3 className="text-xl font-bold text-neutral-900">
                Write with {selected.wordText}
              </h3>
              {canonical.shortDefinition ? (
                <p className="mt-2 text-neutral-700">
                  {canonical.shortDefinition}
                </p>
              ) : null}
              <SaveAndWrite
                key={`${currentUser.id}:${selected.meaningId}`}
                meaningId={selected.meaningId}
                wordSlug={selected.wordSlug}
                wordText={selected.wordText}
                shortDefinition={canonical.shortDefinition}
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
      ) : (
        <Surface>
          <h2 className="text-xl font-bold text-neutral-900">
            Choose a situation to begin
          </h2>
          <p className="mt-3 text-neutral-700">
            Practise a message, a request, or something you want to say in
            everyday life.
          </p>
        </Surface>
      )}
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
