import { DailyReviewsContent } from "../../(app)/reviews/_components/daily-lesson-page";

export { metadata } from "../../(app)/reviews/page";

export default async function ReviewSessionRoute({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const sessionId = typeof query.lesson === "string" ? query.lesson : undefined;
  return <DailyReviewsContent sessionId={sessionId} />;
}
