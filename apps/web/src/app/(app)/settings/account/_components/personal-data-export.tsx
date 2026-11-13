"use client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { Surface } from "@/ui/surface";
import { useState } from "react";

function idempotencyKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
}

export function PersonalDataExport() {
  const [state, setState] = useState<"idle" | "loading" | "done" | "error">(
    "idle",
  );
  const [message, setMessage] = useState("");

  async function download() {
    setState("loading");
    setMessage("");
    try {
      const csrf = await getOrRefreshCSRFToken();
      if (!csrf) {
        setState("error");
        setMessage(
          "We couldn't prepare your export securely. Please try again.",
        );
        return;
      }
      const { data } = await createApiClient().exportPersonalData(
        idempotencyKey(),
        { headers: { "X-CSRF-Token": csrf } },
      );
      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `vocanova-personal-data-${new Date().toISOString().slice(0, 10)}.json`;
      link.click();
      URL.revokeObjectURL(url);
      setState("done");
    } catch (error) {
      setState("error");
      setMessage(
        handleApiError(
          error,
          "We could not create your export. Please try again.",
        ),
      );
    }
  }

  return (
    <Surface
      aria-labelledby="personal-data-export-heading"
      className="mt-[var(--spacing-lg)]"
    >
      <h2
        id="personal-data-export-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        Download your data
      </h2>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Download a JSON copy of your profile, settings, learning activity,
        sentences, and learner-visible AI feedback. It does not include security
        information or hidden AI instructions.
      </p>
      <button
        type="button"
        onClick={download}
        disabled={state === "loading"}
        aria-busy={state === "loading"}
        className="mt-[var(--spacing-md)] vn-btn vn-btn-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
        {state === "loading" ? "Preparing download..." : "Download my data"}
      </button>
      {state === "done" ? (
        <p
          role="status"
          className="mt-[var(--spacing-sm)] text-sm text-green-800"
        >
          Your download has started.
        </p>
      ) : null}
      {state === "error" ? (
        <p role="alert" className="mt-[var(--spacing-sm)] text-sm text-red-800">
          {message}
        </p>
      ) : null}
    </Surface>
  );
}
