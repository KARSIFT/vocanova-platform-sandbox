"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import type {
  CompleteOnboardingBody,
  EnglishLevel,
  LearningGoal,
  MainUseCase,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

interface OnboardingFormProps {
  defaultNativeLanguage?: string;
}

type Status =
  | { type: "idle" }
  | { type: "submitting" }
  | { type: "error"; message: string };

const ENGLISH_LEVELS: { value: EnglishLevel; label: string; helper: string }[] =
  [
    {
      value: "a1",
      label: "A1 — Beginner",
      helper: "I can use very simple phrases.",
    },
    {
      value: "a2",
      label: "A2 — Elementary",
      helper: "I can handle everyday expressions.",
    },
    {
      value: "b1",
      label: "B1 — Intermediate",
      helper: "I can manage most travel situations.",
    },
    {
      value: "b2",
      label: "B2 — Upper-intermediate",
      helper: "I can discuss a range of topics fluently.",
    },
    {
      value: "unknown",
      label: "Not sure yet",
      helper: "I want to start and figure it out.",
    },
  ];

const LEARNING_GOALS: { value: LearningGoal; label: string }[] = [
  { value: "general", label: "General growth" },
  { value: "work", label: "Work" },
  { value: "travel", label: "Travel" },
  { value: "study", label: "Study" },
  { value: "conversation", label: "Conversation" },
  { value: "exam", label: "Exam prep" },
];

const MAIN_USE_CASES: { value: MainUseCase; label: string }[] = [
  { value: "daily_life", label: "Daily life" },
  { value: "work", label: "Work" },
  { value: "travel", label: "Travel" },
  { value: "study", label: "Study" },
  { value: "social", label: "Social" },
];

const DAILY_REVIEW_TARGETS = [5, 10, 15, 20, 30, 50, 75, 100];

const NATIVE_LANGUAGE_SUGGESTIONS = [
  "es",
  "en",
  "pt",
  "fr",
  "de",
  "it",
  "zh",
  "ja",
  "ko",
  "ar",
  "ru",
  "hi",
];

interface FormState {
  englishLevel: EnglishLevel | null;
  nativeLanguage: string;
  learningGoal: LearningGoal | null;
  mainUseCase: MainUseCase | null;
  dailyReviewTarget: number;
}

const INITIAL_STATE: FormState = {
  englishLevel: null,
  nativeLanguage: "",
  learningGoal: null,
  mainUseCase: null,
  dailyReviewTarget: 20,
};

function isFormComplete(state: FormState): boolean {
  return (
    state.englishLevel !== null &&
    state.nativeLanguage.trim().length > 0 &&
    state.learningGoal !== null &&
    state.mainUseCase !== null
  );
}

export function OnboardingForm({
  defaultNativeLanguage = "",
}: OnboardingFormProps) {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [state, setState] = useState<FormState>({
    ...INITIAL_STATE,
    nativeLanguage: defaultNativeLanguage,
  });
  const [status, setStatus] = useState<Status>({ type: "idle" });

  const totalSteps = 5;
  const canGoBack = step > 0 && status.type !== "submitting";
  const canGoForward = useMemo(
    () => stepIsComplete(step, state),
    [step, state],
  );

  function goNext() {
    if (!canGoForward) {
      return;
    }
    setStep((s) => Math.min(totalSteps - 1, s + 1));
  }

  function goBack() {
    if (!canGoBack) {
      return;
    }
    setStep((s) => Math.max(0, s - 1));
  }

  async function handleSubmit() {
    if (!isFormComplete(state)) {
      return;
    }
    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setStatus({
        type: "error",
        message:
          "Your session is missing a security token. Please refresh the page and try again.",
      });
      return;
    }
    setStatus({ type: "submitting" });
    const body: CompleteOnboardingBody = {
      englishLevel: state.englishLevel!,
      nativeLanguage: state.nativeLanguage.trim(),
      learningGoal: state.learningGoal!,
      mainUseCase: state.mainUseCase!,
      dailyReviewTarget: state.dailyReviewTarget,
    };
    const client = createApiClient();
    try {
      await client.completeOnboarding(body, {
        headers: { "X-CSRF-Token": csrfToken },
      });
      // router.push + refresh ensures the Next.js cache for any
      // /me-derived page (e.g. /home) re-reads the additive
      // onboardingStatus field on next render.
      router.push("/home");
      router.refresh();
    } catch (error) {
      // T06: a 401 mid-onboarding means the session expired before the
      // answers were saved. handleApiError routes the learner to
      // re-auth; the per-step answers are still in component state so
      // they survive the page transition. We never claim onboarding is
      // complete when the server rejected the submission.
      const message = handleApiError(
        error,
        "We couldn't save your answers. Please try again.",
      );
      setStatus({ type: "error", message });
    }
  }

  return (
    <div className="space-y-[var(--spacing-lg)]">
      <StepIndicator currentStep={step} totalSteps={totalSteps} />

      {step === 0 ? (
        <EnglishLevelStep
          value={state.englishLevel}
          onChange={(value) => setState((s) => ({ ...s, englishLevel: value }))}
        />
      ) : null}
      {step === 1 ? (
        <NativeLanguageStep
          value={state.nativeLanguage}
          onChange={(value) =>
            setState((s) => ({ ...s, nativeLanguage: value }))
          }
        />
      ) : null}
      {step === 2 ? (
        <LearningGoalStep
          value={state.learningGoal}
          onChange={(value) => setState((s) => ({ ...s, learningGoal: value }))}
        />
      ) : null}
      {step === 3 ? (
        <MainUseCaseStep
          value={state.mainUseCase}
          onChange={(value) => setState((s) => ({ ...s, mainUseCase: value }))}
        />
      ) : null}
      {step === 4 ? (
        <DailyReviewTargetStep
          value={state.dailyReviewTarget}
          onChange={(value) =>
            setState((s) => ({ ...s, dailyReviewTarget: value }))
          }
          state={state}
        />
      ) : null}

      {status.type === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="rounded-md border border-red-300 bg-red-50 p-[var(--spacing-sm)] text-base text-red-800"
        >
          {status.message}
        </p>
      ) : null}

      <div className="flex flex-wrap items-center justify-between gap-[var(--spacing-sm)] pt-[var(--spacing-sm)]">
        <button
          type="button"
          onClick={goBack}
          disabled={!canGoBack}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Back
        </button>

        {step < totalSteps - 1 ? (
          <button
            type="button"
            onClick={goNext}
            disabled={!canGoForward}
            className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Continue
          </button>
        ) : (
          <button
            type="button"
            onClick={handleSubmit}
            disabled={!isFormComplete(state) || status.type === "submitting"}
            aria-busy={status.type === "submitting"}
            className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {status.type === "submitting" ? "Saving..." : "Finish setup"}
          </button>
        )}
      </div>
    </div>
  );
}

function stepIsComplete(step: number, state: FormState): boolean {
  switch (step) {
    case 0:
      return state.englishLevel !== null;
    case 1:
      return state.nativeLanguage.trim().length > 0;
    case 2:
      return state.learningGoal !== null;
    case 3:
      return state.mainUseCase !== null;
    case 4:
      return (
        state.englishLevel !== null &&
        state.nativeLanguage.trim().length > 0 &&
        state.learningGoal !== null &&
        state.mainUseCase !== null
      );
    default:
      return false;
  }
}

function StepIndicator({
  currentStep,
  totalSteps,
}: {
  currentStep: number;
  totalSteps: number;
}) {
  return (
    <ol
      aria-label="Onboarding progress"
      className="flex w-full items-center gap-[var(--spacing-xs)]"
    >
      {Array.from({ length: totalSteps }).map((_, index) => {
        const isComplete = index < currentStep;
        const isCurrent = index === currentStep;
        return (
          <li
            key={index}
            aria-current={isCurrent ? "step" : undefined}
            className={`h-2 flex-1 rounded-full ${
              isComplete
                ? "bg-primary-600"
                : isCurrent
                  ? "bg-primary-300"
                  : "bg-neutral-200"
            }`}
            data-state={
              isComplete ? "complete" : isCurrent ? "current" : "upcoming"
            }
          >
            <span className="sr-only">
              {isComplete
                ? `Step ${index + 1} of ${totalSteps} complete`
                : isCurrent
                  ? `Step ${index + 1} of ${totalSteps} (current)`
                  : `Step ${index + 1} of ${totalSteps} (upcoming)`}
            </span>
          </li>
        );
      })}
    </ol>
  );
}

interface RadioStepProps<T extends string> {
  legend: string;
  description: string;
  name: string;
  options: { value: T; label: string; helper?: string }[];
  value: T | null;
  onChange: (next: T) => void;
}

function RadioStep<T extends string>({
  legend,
  description,
  name,
  options,
  value,
  onChange,
}: RadioStepProps<T>) {
  return (
    <fieldset className="space-y-[var(--spacing-md)]">
      <legend className="text-xl font-semibold text-neutral-900">
        {legend}
      </legend>
      <p className="text-base text-neutral-700">{description}</p>
      <div
        role="radiogroup"
        aria-label={legend}
        className="space-y-[var(--spacing-sm)]"
      >
        {options.map((option) => {
          const checked = value === option.value;
          return (
            <label
              key={option.value}
              className={`flex min-h-[var(--spacing-2xl)] cursor-pointer items-start gap-[var(--spacing-sm)] rounded-md border p-[var(--spacing-md)] transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] ${
                checked
                  ? "border-primary-600 bg-primary-50"
                  : "border-neutral-200 bg-white hover:border-primary-300"
              } focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700`}
            >
              <input
                type="radio"
                name={name}
                value={option.value}
                checked={checked}
                onChange={() => onChange(option.value)}
                className="mt-[var(--spacing-xs)] size-4 accent-primary-600"
              />
              <span className="flex flex-col">
                <span className="text-base font-medium text-neutral-900">
                  {option.label}
                </span>
                {option.helper ? (
                  <span className="text-sm text-neutral-700">
                    {option.helper}
                  </span>
                ) : null}
              </span>
            </label>
          );
        })}
      </div>
    </fieldset>
  );
}

function EnglishLevelStep({
  value,
  onChange,
}: {
  value: EnglishLevel | null;
  onChange: (next: EnglishLevel) => void;
}) {
  return (
    <RadioStep
      legend="How would you describe your English?"
      description="Choose the closest level for now. Later, Settings lets you adjust your daily review target."
      name="englishLevel"
      value={value}
      onChange={onChange}
      options={ENGLISH_LEVELS}
    />
  );
}

function NativeLanguageStep({
  value,
  onChange,
}: {
  value: string;
  onChange: (next: string) => void;
}) {
  return (
    <div className="space-y-[var(--spacing-md)]">
      <div>
        <h2 className="text-xl font-semibold text-neutral-900">
          What&apos;s your native language?
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          We use this to know when a word or example is in your language.
        </p>
      </div>
      <label
        className="block text-base font-medium text-neutral-900"
        htmlFor="native-language"
      >
        Native language
      </label>
      <input
        id="native-language"
        name="nativeLanguage"
        type="text"
        autoComplete="off"
        required
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        aria-describedby="native-language-helper native-language-suggestions"
      />
      <p id="native-language-helper" className="text-sm text-neutral-700">
        Use the short form (for example, &ldquo;Spanish&rdquo; or
        &ldquo;es&rdquo;).
      </p>
      <div
        id="native-language-suggestions"
        className="flex flex-wrap gap-[var(--spacing-xs)]"
        aria-label="Common suggestions"
      >
        {NATIVE_LANGUAGE_SUGGESTIONS.map((suggestion) => {
          const isSelected = value.trim().toLowerCase() === suggestion;
          return (
            <button
              key={suggestion}
              type="button"
              onClick={() => onChange(suggestion)}
              aria-pressed={isSelected}
              className={`min-h-[var(--spacing-xl)] rounded-full border px-[var(--spacing-sm)] text-sm font-medium transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 ${
                isSelected
                  ? "border-primary-600 bg-primary-50 text-primary-800"
                  : "border-neutral-300 bg-white text-neutral-800 hover:border-primary-300"
              }`}
            >
              {suggestion}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function LearningGoalStep({
  value,
  onChange,
}: {
  value: LearningGoal | null;
  onChange: (next: LearningGoal) => void;
}) {
  return (
    <RadioStep
      legend="What's your main reason for learning?"
      description="Pick the one that fits best right now."
      name="learningGoal"
      value={value}
      onChange={onChange}
      options={LEARNING_GOALS}
    />
  );
}

function MainUseCaseStep({
  value,
  onChange,
}: {
  value: MainUseCase | null;
  onChange: (next: MainUseCase) => void;
}) {
  return (
    <RadioStep
      legend="Where will you use English most?"
      description="We'll lean your word choices toward this setting."
      name="mainUseCase"
      value={value}
      onChange={onChange}
      options={MAIN_USE_CASES}
    />
  );
}

function DailyReviewTargetStep({
  value,
  onChange,
  state,
}: {
  value: number;
  onChange: (next: number) => void;
  state: FormState;
}) {
  return (
    <div className="space-y-[var(--spacing-md)]">
      <div>
        <h2 className="text-xl font-semibold text-neutral-900">
          Daily review target
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          We&apos;ll suggest {value} words a day for review. You can change this
          in Settings any time.
        </p>
      </div>
      <div
        role="radiogroup"
        aria-label="Daily review target"
        className="grid grid-cols-2 gap-[var(--spacing-sm)] sm:grid-cols-4"
      >
        {DAILY_REVIEW_TARGETS.map((target) => {
          const checked = value === target;
          return (
            <div key={target}>
              <label
                className={`flex min-h-[var(--spacing-2xl)] cursor-pointer items-center justify-center rounded-md border p-[var(--spacing-sm)] text-base font-medium transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${
                  checked
                    ? "border-primary-600 bg-primary-50 text-primary-800"
                    : "border-neutral-200 bg-white text-neutral-900 hover:border-primary-300"
                }`}
              >
                <input
                  type="radio"
                  name="dailyReviewTarget"
                  value={target}
                  checked={checked}
                  onChange={() => onChange(target)}
                  className="sr-only"
                />
                {target}
              </label>
            </div>
          );
        })}
      </div>
      <SummaryCard state={state} dailyReviewTarget={value} />
    </div>
  );
}

function SummaryCard({
  state,
  dailyReviewTarget,
}: {
  state: FormState;
  dailyReviewTarget: number;
}) {
  const englishLabel =
    ENGLISH_LEVELS.find((level) => level.value === state.englishLevel)?.label ??
    "Not set";
  const goalLabel =
    LEARNING_GOALS.find((goal) => goal.value === state.learningGoal)?.label ??
    "Not set";
  const useLabel =
    MAIN_USE_CASES.find((use) => use.value === state.mainUseCase)?.label ??
    "Not set";
  return (
    <aside
      aria-label="Onboarding summary"
      className="rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] text-base text-neutral-800 shadow-sm"
    >
      <h3 className="text-lg font-semibold text-neutral-900">Quick check</h3>
      <dl className="mt-[var(--spacing-sm)] space-y-[var(--spacing-xs)]">
        <div className="flex justify-between">
          <dt className="text-neutral-600">English level</dt>
          <dd className="font-medium text-neutral-900">{englishLabel}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-neutral-600">Native language</dt>
          <dd className="font-medium text-neutral-900">
            {state.nativeLanguage.trim() || "Not set"}
          </dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-neutral-600">Reason for learning</dt>
          <dd className="font-medium text-neutral-900">{goalLabel}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-neutral-600">Main use case</dt>
          <dd className="font-medium text-neutral-900">{useLabel}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-neutral-600">Daily reviews</dt>
          <dd className="font-medium text-neutral-900">{dailyReviewTarget}</dd>
        </div>
      </dl>
    </aside>
  );
}
