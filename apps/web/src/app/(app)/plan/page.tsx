import type { Metadata } from "next";
import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { Eyebrow, PageContainer, Surface } from "@/ui/surface";
import {
  learningGoals as goals,
  learningFocuses as focuses,
} from "@/lib/learning-direction";
import { LearningDirection } from "./_components/learning-direction";
import { RecommendedLesson } from "../_components/lesson-recommendation";

export const metadata: Metadata = { title: "Your learning plan — Vocanova" };
const action =
  "mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const link =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

async function optional<T>(request: Promise<T>): Promise<T | null> {
  try {
    return await request;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, "/plan");
    return null;
  }
}

export default async function LearningPlanPage() {
  const client = await createServerApiClient();
  const [
    profileResponse,
    settingsResponse,
    missionResponse,
    dueResponse,
    recommendationResponse,
  ] = await Promise.all([
    optional(client.getLearningPreferences()),
    optional(client.getSettings()),
    optional(client.getDailyMission()),
    optional(client.listDueWords({ limit: 1 })),
    optional(client.getLessonRecommendation()),
  ]);
  const profile = profileResponse?.data;
  const settings = settingsResponse?.data;
  const mission = missionResponse?.data;
  const dueCount = dueResponse?.data.totalCount;
  const targetComplete = mission
    ? mission.reviewsCompleted >= mission.reviewTarget
    : false;
  const reviewReady =
    dueCount !== undefined && dueCount > 0 && mission && !targetComplete;

  return (
    <PageContainer className="max-w-[64rem]">
      <Link href="/home" className={link}>
        Back to Home
      </Link>
      <header className="mb-6 mt-3">
        <Eyebrow>A plan you can make your own</Eyebrow>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-neutral-900">
          Your learning plan
        </h1>
        <p className="mt-3 max-w-[40rem] text-lg text-neutral-700">
          Start with useful words, practise a little, and return to them when
          they are due.
        </p>
      </header>
      <Surface aria-labelledby="plan-preferences-heading" tone="primary">
        <h2
          id="plan-preferences-heading"
          className="text-xl font-bold text-neutral-900"
        >
          What you are working towards
        </h2>
        <dl className="mt-4 grid gap-4 sm:grid-cols-3">
          <div>
            <dt className="text-sm text-neutral-600">Your goal</dt>
            <dd className="mt-1 font-semibold text-neutral-900">
              {profile?.learningGoal
                ? goals[profile.learningGoal]
                : profile
                  ? "Not chosen yet"
                  : "Goal unavailable"}
            </dd>
          </div>
          <div>
            <dt className="text-sm text-neutral-600">Your focus</dt>
            <dd className="mt-1 font-semibold text-neutral-900">
              {profile?.mainUseCase
                ? focuses[profile.mainUseCase]
                : profile
                  ? "Not chosen yet"
                  : "Focus unavailable"}
            </dd>
          </div>
          <div>
            <dt className="text-sm text-neutral-600">Your daily pace</dt>
            <dd className="mt-1 font-semibold text-neutral-900">
              {settings
                ? `${settings.dailyReviewTarget} reviews a day`
                : "Daily pace unavailable"}
            </dd>
          </div>
        </dl>
        <p className="mt-3 text-sm text-neutral-600">
          Your focus, known words and review progress guide lesson suggestions.
          Keep your personal goal here and change direction whenever your needs
          change.
        </p>
        <LearningDirection initial={profile ?? null} />
        <Link href="/settings" className={`${link} mt-3`}>
          Change daily preferences
        </Link>
        <Link
          href="/settings#practice-reminder"
          className={`${link} ml-4 mt-3`}
        >
          Set a calendar reminder
        </Link>
      </Surface>
      <div className="mt-4 grid gap-4 md:grid-cols-2">
        <Surface aria-labelledby="plan-review-heading">
          <Eyebrow>Keep words fresh</Eyebrow>
          <h2
            id="plan-review-heading"
            className="mt-2 text-xl font-bold text-neutral-900"
          >
            Your reviews today
          </h2>
          {mission ? (
            <p className="mt-3 font-semibold text-neutral-900">
              {mission.reviewsCompleted} of {mission.reviewTarget} reviews
              complete today.
            </p>
          ) : (
            <p role="status" className="mt-3 text-neutral-700">
              Your daily progress is unavailable right now.
            </p>
          )}
          <p className="mt-2 text-neutral-700">
            {dueCount === undefined
              ? "Open your reviews to check what is ready."
              : `${dueCount} ${dueCount === 1 ? "word is" : "words are"} due now.`}
          </p>
          <p className="mt-2 text-sm text-neutral-600">
            {targetComplete
              ? "Today’s target is complete. Come back tomorrow, or choose another activity."
              : "Work with words that are due, up to your daily target. You do not need to add more just to fill a session."}
          </p>
          {settings &&
            mission &&
            settings.dailyReviewTarget !== mission.reviewTarget && (
              <p className="mt-2 text-sm text-neutral-600">
                Your updated pace is {settings.dailyReviewTarget} reviews a day.
                Today’s existing target remains {mission.reviewTarget}.
              </p>
            )}
          <Link href="/review" className={action}>
            {reviewReady ? "Start review" : "View reviews"}
          </Link>
        </Surface>
        <RecommendedLesson data={recommendationResponse?.data ?? null} />
      </div>
      <Surface aria-labelledby="plan-words-heading" className="mt-4">
        <h2
          id="plan-words-heading"
          className="text-xl font-bold text-neutral-900"
        >
          Make room for words you need
        </h2>
        <p className="mt-3 max-w-[40rem] text-neutral-700">
          Use the optional self-check to mark meanings you already know and save
          the ones you want to learn. You can change these choices on each
          word’s page.
        </p>
        <p className="mt-2 text-sm text-neutral-600">
          “Already known” is your own assessment. Saved words, review progress
          and practice results remain separate.
        </p>
        <div className="mt-4 flex flex-wrap gap-4">
          <Link href="/vocabulary/check" className={link}>
            Find your starting words
          </Link>
          <Link href="/vocabulary?view=map" className={link}>
            Open your vocabulary map
          </Link>
          <Link href="/practice" className={link}>
            Choose a practice
          </Link>
        </div>
      </Surface>
    </PageContainer>
  );
}
