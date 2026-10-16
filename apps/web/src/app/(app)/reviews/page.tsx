import type { Metadata } from "next";
import { DailyReviewsContent } from "./_components/daily-lesson-page";

export const metadata: Metadata = {
  title: "Today’s lesson — Vocanova",
  description: "Remember today’s words, one short question at a time.",
};

export default function ReviewsPage() {
  return <DailyReviewsContent />;
}
