import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";

import { getDiscoverListView } from "./_components/discover-view";
import { formatLevelBand } from "./_components/level-band";
import { LessonPath } from "../_components/lesson-path";

export default async function DiscoverPage() {
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.listJourneySituations>>;
  try {
    response = await client.listJourneySituations();
  } catch (error) {
    requireAuthRedirect(error, "/discover");
  }

  const { items } = response.data;
  const lessonResponse = await client.listLessons().catch(() => null);
  const situationOrder = new Map(
    items.map((item, index) => [item.slug, index] as const),
  );
  const lessons = [...(lessonResponse?.data.items ?? [])].sort(
    (a, b) =>
      (situationOrder.get(a.situationSlug) ?? 999) -
      (situationOrder.get(b.situationSlug) ?? 999),
  );

  return (
    <PageContainer className="max-w-[64rem]">
      <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-neutral-900">
          Journey
        </h1>
        <div className="flex flex-wrap gap-5">
          <Link
            href="/vocabulary"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
          >
            Find a word
          </Link>
          <Link
            href="/words"
            className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
          >
            Saved words
          </Link>
        </div>
      </div>
      {lessonResponse ? (
        <LessonPath lessons={lessons} />
      ) : (
        <p role="status" className="my-6 text-neutral-700">
          Guided lessons are unavailable right now. You can still explore the
          situations below.
        </p>
      )}
      <details className="my-6 border-y border-neutral-200 py-2">
        <summary className="min-h-11 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
          More ways to practise
        </summary>
        <nav
          aria-label="More ways to practise"
          className="grid gap-2 py-3 sm:grid-cols-2"
        >
          {(
            [
              ["/practice", "Choose a practice"],
              ["/writing", "Topic writing"],
              ["/lists", "Personal lists"],
              ["/plan", "Your learning plan"],
            ] as const
          ).map(([href, label]) => (
            <Link
              key={href}
              href={href}
              className="inline-flex min-h-11 items-center rounded-xl bg-white px-4 py-3 font-semibold text-primary-700 hover:bg-primary-50"
            >
              {label}
            </Link>
          ))}
        </nav>
      </details>
      <h2 className="mt-8 text-2xl font-bold tracking-tight text-neutral-900">
        Explore by situation
      </h2>

      {getDiscoverListView(items.length) === "empty" ? (
        <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            New situations are on the way
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            We don&apos;t have any situations to explore just yet. Check back
            soon, or head to your reviews in the meantime.
          </p>
          <Link
            href="/review"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Go to Reviews
          </Link>
        </div>
      ) : (
        <ul className="mt-[var(--spacing-lg)] grid grid-cols-1 gap-[var(--spacing-md)] sm:grid-cols-2">
          {items.map((situation) => (
            <li key={situation.slug}>
              <Link
                href={`/discover/${situation.slug}`}
                className="group block rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm transition-colors hover:border-primary-300 hover:shadow-md"
              >
                <div className="flex items-center gap-[var(--spacing-sm)]">
                  <span className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-secondary-100 text-secondary-800">
                    <SituationIcon
                      slug={situation.slug}
                      category={situation.category}
                    />
                  </span>
                  <div className="min-w-0">
                    <h2 className="text-lg font-bold tracking-tight text-neutral-900 group-hover:text-primary-800">
                      {situation.title}
                    </h2>
                    <span className="mt-1 inline-flex rounded-lg bg-secondary-100 px-[var(--spacing-sm)] py-1 text-xs font-bold text-secondary-800">
                      {situation.category.replaceAll("_", " ")}
                      {situation.levelBand
                        ? ` · ${formatLevelBand(situation.levelBand)}`
                        : ""}
                    </span>
                  </div>
                </div>
                <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
                  {situation.shortDescription}
                </p>
                <span className="mt-[var(--spacing-xs)] inline-flex min-h-11 items-center text-sm font-semibold text-primary-700 group-hover:text-primary-800">
                  Explore words
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </PageContainer>
  );
}

function SituationIcon({
  slug,
  category,
}: Readonly<{ slug: string; category: string }>) {
  const sharedProps = {
    "aria-hidden": true,
    viewBox: "0 0 24 24",
    className: "h-5 w-5 fill-none stroke-current stroke-[1.8]",
  } as const;

  if (slug.includes("cafe") || slug.includes("restaurant")) {
    return (
      <svg {...sharedProps}>
        <path
          strokeLinecap="round"
          d="M5 8h10v5a4 4 0 0 1-8 0V8Zm10 2h2a2 2 0 0 1 0 4h-2M7 4v2m3-2v2m3-2v2M4 20h14"
        />
      </svg>
    );
  }
  if (slug.includes("hotel")) {
    return (
      <svg {...sharedProps}>
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M4 19V7h16v12M7 7V4h10v3m-9 5h3m3 0h2M3 19h18"
        />
      </svg>
    );
  }
  if (slug.includes("airport") || category === "travel") {
    return (
      <svg {...sharedProps}>
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="m3.5 14.5 17-3.5-17-3.5 4.5 3.5-4.5 3.5Zm5 1.5 2 2.5h3l1-3"
        />
      </svg>
    );
  }
  if (slug.includes("meeting") || slug.includes("interview")) {
    return (
      <svg {...sharedProps}>
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M5 19V8h14v11M8 8V5h8v3m-7 5h6m-6 3h3"
        />
      </svg>
    );
  }
  if (category === "social" || slug.includes("social")) {
    return (
      <svg {...sharedProps}>
        <circle cx="9" cy="9" r="2.5" />
        <circle cx="16" cy="10" r="2" />
        <path
          strokeLinecap="round"
          d="M4.5 19c.4-3 2.1-4.5 4.5-4.5s4.1 1.5 4.5 4.5m0-2c.5-1.5 1.6-2.3 3.2-2.3 1.6 0 2.5.8 2.8 2.3"
        />
      </svg>
    );
  }
  if (category === "work") {
    return (
      <svg {...sharedProps}>
        <rect x="4" y="7" width="16" height="12" rx="2" />
        <path strokeLinecap="round" d="M9 7V5.5h6V7m-11 5h16m-8 0v2" />
      </svg>
    );
  }
  return (
    <svg {...sharedProps}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M5 8h14v11H5zM8 5h8v3m-5 5h2"
      />
      <path strokeLinecap="round" d="M3.5 11h17" />
    </svg>
  );
}
