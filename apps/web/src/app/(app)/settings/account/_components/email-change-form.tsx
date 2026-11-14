"use client";

import { StatusMessage } from "@/ui/status-message";

import { Button, TextInput } from "@/ui/controls";

import { useEffect, useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import {
  clearEmailChangeDraft,
  isPreferencesDraftDocumentActive,
  readEmailChangeDraft,
  saveEmailChangeDraft,
} from "@/lib/preferences-draft";

interface EmailChangeFormProps {
  currentEmail: string;
  userId?: string;
}

type EmailPhase =
  | { type: "idle" }
  | { type: "requesting" }
  | { type: "pending"; newEmail: string }
  | { type: "error"; message: string };

export function EmailChangeForm({
  currentEmail,
  userId,
}: EmailChangeFormProps) {
  const [newEmail, setNewEmail] = useState("");
  const [phase, setPhase] = useState<EmailPhase>({ type: "idle" });
  const [restored, setRestored] = useState(false);
  const requesting = useRef(false);
  const mounted = useRef(false);
  const recoveryScope = useRef<{
    userId?: string;
    currentEmail: string;
  } | null>(null);
  const emailInput = useRef<HTMLInputElement>(null);
  const focusEmailOnReturn = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (
      recoveryScope.current?.userId === userId &&
      recoveryScope.current?.currentEmail === currentEmail
    )
      return;
    recoveryScope.current = { userId, currentEmail };
    const draft = readEmailChangeDraft(userId, currentEmail);
    setNewEmail(draft ?? "");
    setRestored(draft !== null);
    setPhase({ type: "idle" });
  }, [currentEmail, userId]);

  useEffect(() => {
    if (phase.type === "idle" && focusEmailOnReturn.current) {
      emailInput.current?.focus();
      focusEmailOnReturn.current = false;
    }
  }, [phase.type]);

  async function handleRequest(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (requesting.current || !isPreferencesDraftDocumentActive()) return;
    const trimmed = newEmail.trim();
    if (!trimmed) {
      return;
    }

    requesting.current = true;
    const scope = recoveryScope.current;
    saveEmailChangeDraft({ userId, currentEmail, newEmail });
    setPhase({ type: "requesting" });
    const client = createApiClient();
    try {
      const csrfToken = await getOrRefreshCSRFToken();
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      if (!csrfToken) {
        setPhase({
          type: "error",
          message:
            "We couldn't prepare this request securely. Please try again.",
        });
        return;
      }
      await client.requestEmailChangeLink(
        { newEmail: trimmed },
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      clearEmailChangeDraft(userId);
      setRestored(false);
      setPhase({ type: "pending", newEmail: trimmed });
    } catch (error) {
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      // Preserve only the unsent input before a 401 leaves this document.
      // The confirmation proof and pending/sent phase are never persisted.
      saveEmailChangeDraft({ userId, currentEmail, newEmail });
      const message = handleApiError(
        error,
        "We couldn't send the confirmation email. Please try again.",
      );
      setPhase({ type: "error", message });
    } finally {
      requesting.current = false;
    }
  }

  function handleStartOver() {
    clearEmailChangeDraft(userId);
    setRestored(false);
    focusEmailOnReturn.current = true;
    setPhase({ type: "idle" });
    setNewEmail("");
  }

  return (
    <div className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]">
      {restored ? (
        <StatusMessage as="p" tone="info" role="status" className="text-sm">
          Your unsent email address was restored in this tab. Check it before
          sending a confirmation link.
        </StatusMessage>
      ) : null}
      {phase.type === "pending" ? (
        <div className="space-y-[var(--spacing-md)]">
          <StatusMessage
            as="p"
            tone="info"
            role="status"
            aria-live="polite"
            className="rounded-md border border-primary-300 bg-primary-50 p-[var(--spacing-md)] text-base"
          >
            If this address can be used, we&apos;ve sent a confirmation link to{" "}
            <span className="font-medium [overflow-wrap:anywhere]">
              {phase.newEmail}
            </span>
            . The link expires in 15 minutes. Open it in this browser while you
            are signed in to finish the change.
          </StatusMessage>
          <Button
            type="button"
            onClick={handleStartOver}
            variant="secondary"
            className="min-w-[var(--spacing-2xl)]"
          >
            Use a different address
          </Button>
        </div>
      ) : null}

      {phase.type === "idle" ||
      phase.type === "requesting" ||
      phase.type === "error" ? (
        <form
          onSubmit={handleRequest}
          aria-label="Request new sign-in email"
          className="space-y-[var(--spacing-sm)]"
        >
          <label
            htmlFor="new-email"
            className="block text-base font-medium text-neutral-900"
          >
            New sign-in email
          </label>
          <TextInput
            id="new-email"
            ref={emailInput}
            name="newEmail"
            type="email"
            autoComplete="email"
            maxLength={320}
            required
            value={newEmail}
            onChange={(event) => {
              setNewEmail(event.target.value);
              saveEmailChangeDraft({
                userId,
                currentEmail,
                newEmail: event.target.value,
              });
              if (!event.target.value.trim()) setRestored(false);
              if (phase.type === "error") {
                setPhase({ type: "idle" });
              }
            }}
            disabled={phase.type === "requesting"}
            placeholder="you@example.com"
            aria-describedby="new-email-helper"
            className="w-full"
          />
          <p id="new-email-helper" className="text-sm text-neutral-700">
            We&apos;ll send a confirmation link to this address. The response is
            the same whether this address is already registered or not, to
            protect learner privacy.
          </p>
          <Button
            type="submit"
            disabled={
              phase.type === "requesting" || newEmail.trim().length === 0
            }
            aria-busy={phase.type === "requesting"}
          >
            {phase.type === "requesting"
              ? "Sending confirmation link..."
              : "Send confirmation link"}
          </Button>
        </form>
      ) : null}

      {phase.type === "error" ? (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          aria-live="assertive"
          className="rounded-md border border-red-300 bg-red-50 p-[var(--spacing-sm)] text-base"
        >
          {phase.message}
        </StatusMessage>
      ) : null}

      {currentEmail ? (
        <p className="text-sm text-neutral-700">
          Your current sign-in address is{" "}
          <span className="break-all font-medium text-neutral-900">
            {currentEmail}
          </span>
          .
        </p>
      ) : null}
    </div>
  );
}
