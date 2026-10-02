"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
  ApiResponseError,
  type LearningGoal,
  type MainUseCase,
  type LearningPreferences,
  type LearningPreferencesUpdate,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { learningGoals, learningFocuses } from "@/lib/learning-direction";

const action =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 disabled:opacity-60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const input =
  "mt-2 min-h-12 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

export function LearningDirection({
  initial,
}: {
  initial: LearningPreferences | null;
}) {
  const router = useRouter();
  const [baseline, setBaseline] = useState(initial);
  const [goal, setGoal] = useState<LearningGoal | "">(
    initial?.learningGoal ?? "",
  );
  const [focus, setFocus] = useState<MainUseCase | "">(
    initial?.mainUseCase ?? "",
  );
  const [open, setOpen] = useState(
    !initial?.learningGoal || !initial?.mainUseCase,
  );
  const [busy, setBusy] = useState(false);
  const [retry, setRetry] = useState(false);
  const [reload, setReload] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const pending = useRef<LearningPreferencesUpdate | null>(null);
  const inFlight = useRef(false);
  const locked = busy || retry || reload;

  async function loadCurrent() {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const { data } = await createApiClient().getLearningPreferences();
      setBaseline(data);
      setGoal(data.learningGoal ?? "");
      setFocus(data.mainUseCase ?? "");
      pending.current = null;
      setRetry(false);
      setReload(false);
      setNotice("Current choices loaded. Review them before saving.");
      router.refresh();
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not load your learning direction. Try again.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function save() {
    if (inFlight.current || reload || !baseline) return;
    if (!pending.current) {
      if (!goal || !focus) return;
      pending.current = {
        learningGoal: goal,
        mainUseCase: focus,
        expectedRevision: baseline.revision,
      };
    }
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setNotice("");
    setError("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().updateLearningPreferences(
        request,
        { headers: { "X-CSRF-Token": token } },
      );
      if (
        data.learningGoal !== request.learningGoal ||
        data.mainUseCase !== request.mainUseCase
      ) {
        setReload(true);
        setRetry(false);
        setError(
          "Your choices changed elsewhere. Load your current choices before saving again.",
        );
        return;
      }
      setBaseline(data);
      pending.current = null;
      setRetry(false);
      setNotice("Your learning direction is saved.");
      router.refresh();
    } catch (cause) {
      if (cause instanceof ApiResponseError && cause.status === 409) {
        setReload(true);
        setRetry(false);
        setError(
          "Your choices changed elsewhere. Load your current choices before saving again.",
        );
      } else {
        setRetry(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm your choices. Retry the same change safely.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  if (!open)
    return (
      <button
        type="button"
        className={`${action} mt-4`}
        onClick={() => setOpen(true)}
      >
        Change learning direction
      </button>
    );

  return (
    <form
      aria-label="Learning direction"
      className="mt-5 border-t border-primary-200 pt-5"
      onSubmit={(event) => {
        event.preventDefault();
        void save();
      }}
    >
      <h3 className="text-lg font-bold text-neutral-900">
        Choose your direction
      </h3>
      <p className="mt-2 text-sm text-neutral-700">
        Your focus guides suggested situations and starting words. All lessons
        remain available.
      </p>
      {baseline ? (
        <>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <label className="font-semibold text-neutral-900">
              Learning goal
              <select
                value={goal}
                disabled={locked}
                required
                className={input}
                onChange={(event) => {
                  setGoal(event.target.value as LearningGoal);
                  setNotice("");
                }}
              >
                <option value="" disabled>
                  Choose a goal
                </option>
                {Object.entries(learningGoals).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            <label className="font-semibold text-neutral-900">
              Main focus
              <select
                value={focus}
                disabled={locked}
                required
                className={input}
                onChange={(event) => {
                  setFocus(event.target.value as MainUseCase);
                  setNotice("");
                }}
              >
                <option value="" disabled>
                  Choose a focus
                </option>
                {Object.entries(learningFocuses).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
          </div>
          {!retry && !reload && (
            <button
              type="submit"
              className={`${action} mt-4`}
              disabled={
                busy ||
                !goal ||
                !focus ||
                (goal === baseline.learningGoal &&
                  focus === baseline.mainUseCase)
              }
            >
              Save learning direction
            </button>
          )}
        </>
      ) : (
        <p role="status" className="mt-3 text-neutral-700">
          Load your current choices to change your direction.
        </p>
      )}
      {busy && (
        <p role="status" className="mt-3 text-neutral-700">
          Saving or loading your choices…
        </p>
      )}
      {error && (
        <p role="alert" className="mt-3 text-neutral-900">
          {error}
        </p>
      )}
      {notice && (
        <p role="status" className="mt-3 text-neutral-700">
          {notice}
        </p>
      )}
      {retry && (
        <button
          type="button"
          disabled={busy}
          className={`${action} mt-3`}
          onClick={() => void save()}
        >
          Retry saving direction
        </button>
      )}
      {(reload || !baseline) && (
        <button
          type="button"
          disabled={busy}
          className={`${action} mt-3`}
          onClick={() => void loadCurrent()}
        >
          Load current choices
        </button>
      )}
    </form>
  );
}
