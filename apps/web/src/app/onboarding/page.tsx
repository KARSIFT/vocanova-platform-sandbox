import { redirect } from "next/navigation";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

import { OnboardingForm } from "./_components/onboarding-form";

export const metadata = {
  title: "Welcome — Vocanova",
  description:
    "A few quick questions so we can shape Vocanova around your English goals.",
};

export default async function OnboardingPage() {
  const client = await createServerApiClient();
  let meResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    meResponse = await client.getCurrentUser();
  } catch (error) {
    requireAuthRedirect(error, "/onboarding");
  }

  // T01 gating: a learner whose onboarding is already completed is
  // redirected away from /onboarding to /home so the form never
  // re-opens for an already-finished account.
  if (meResponse.data.onboardingStatus === "completed") {
    redirect("/home");
  }

  return (
    <main className="grid min-h-screen place-items-center bg-neutral-50 p-6">
      <div className="w-full max-w-xl space-y-[var(--spacing-lg)] rounded-xl border border-neutral-200 bg-white p-[var(--spacing-lg)] shadow-sm">
        <header className="space-y-[var(--spacing-xs)]">
          <h1 className="text-2xl font-semibold text-neutral-900">
            Welcome to Vocanova
          </h1>
          <p className="text-base text-neutral-700">
            A few quick questions so we can shape your practice. You can change
            every answer later in Settings.
          </p>
        </header>
        <OnboardingForm />
      </div>
    </main>
  );
}
