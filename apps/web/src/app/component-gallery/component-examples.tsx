"use client";

import { useRef, useState, type FormEvent } from "react";
import { useThemePreference } from "../_components/theme-preference";
import { isThemePreference } from "@/lib/theme-preference";
import { Button, ActionLink, TextInput, TextArea, Select } from "@/ui/controls";
import { Surface } from "@/ui/surface";
import { LearningIntro, LearningNote } from "@/ui/learning-guide";
import { StatusMessage } from "@/ui/status-message";

export function ComponentExamples() {
  const { preference, setPreference } = useThemePreference();
  const [submitted, setSubmitted] = useState(0);
  const [message, setMessage] = useState("");
  const [error, setError] = useState(false);
  const [pending, setPending] = useState(false);
  const sentence = useRef<HTMLTextAreaElement>(null);
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!sentence.current?.value.trim()) {
      setError(true);
      sentence.current?.focus();
      return;
    }
    setError(false);
    setSubmitted((count) => count + 1);
    setMessage(
      "Your example was checked in this preview. No learning record was saved.",
    );
  }
  return (
    <div className="mt-8 space-y-6">
      <Surface aria-labelledby="preview-settings">
        <h2 id="preview-settings" className="text-xl font-semibold">
          Preview settings
        </h2>
        <label htmlFor="gallery-theme" className="mt-4 block font-semibold">
          Appearance
        </label>
        <Select
          id="gallery-theme"
          value={preference}
          onChange={(event) => {
            if (isThemePreference(event.target.value))
              setPreference(event.target.value);
          }}
        >
          <option value="system">System</option>
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </Select>
        <p className="mt-3 text-neutral-700">
          Compare this page at 360px, 430px and desktop, including enlarged text
          and keyboard navigation.
        </p>
      </Surface>
      <Surface aria-labelledby="gallery-actions">
        <h2 id="gallery-actions" className="text-xl font-semibold">
          Actions
        </h2>
        <div className="mt-4 flex flex-wrap gap-3">
          <Button onClick={() => setMessage("Preview action completed.")}>
            Primary action
          </Button>
          <Button variant="secondary" onClick={() => sentence.current?.focus()}>
            Focus sentence
          </Button>
          <Button variant="text" onClick={() => setMessage("")}>
            Clear message
          </Button>
          <Button
            variant="danger"
            onClick={() => setMessage("Only the preview message was cleared.")}
          >
            Remove example
          </Button>
          <Button disabled>Unavailable action</Button>
          <Button disabled aria-busy="true">
            Checking example…
          </Button>
          <ActionLink href="#gallery-form" variant="secondary">
            Go to form example
          </ActionLink>
        </div>
        <div className="mt-4">
          <StatusMessage tone="success" aria-label="Preview action status">
            {message}
          </StatusMessage>
        </div>
      </Surface>
      <Surface aria-labelledby="gallery-form">
        <h2 id="gallery-form" className="text-xl font-semibold">
          Form fields
        </h2>
        <form className="mt-4 space-y-4" onSubmit={submit}>
          <div>
            <label htmlFor="gallery-word" className="block font-semibold">
              Word
            </label>
            <TextInput
              id="gallery-word"
              name="word"
              defaultValue="invite"
              required
            />
          </div>
          <div>
            <label htmlFor="gallery-mode" className="block font-semibold">
              Practice mode
            </label>
            <Select id="gallery-mode" name="mode">
              <option>Writing</option>
              <option>Meaning recall</option>
            </Select>
          </div>
          <div>
            <label htmlFor="gallery-sentence" className="block font-semibold">
              Your sentence
            </label>
            <TextArea
              ref={sentence}
              id="gallery-sentence"
              name="sentence"
              rows={3}
              aria-invalid={error || undefined}
              aria-describedby={
                error ? "gallery-sentence-error" : "gallery-sentence-help"
              }
            />
            <p id="gallery-sentence-help" className="mt-2 text-neutral-700">
              Use invite in a sentence about your day.
            </p>
            {error && (
              <StatusMessage tone="error" id="gallery-sentence-error">
                Write a sentence before checking.
              </StatusMessage>
            )}
          </div>
          <div>
            <label htmlFor="gallery-disabled" className="block font-semibold">
              Unavailable field
            </label>
            <TextInput
              id="gallery-disabled"
              defaultValue="Waiting for an activity"
              disabled
            />
          </div>
          <div className="flex flex-wrap gap-3">
            <Button type="submit" disabled={pending}>
              Check preview sentence
            </Button>
            <Button
              variant="secondary"
              onClick={() => setPending((value) => !value)}
            >
              {pending ? "End simulated wait" : "Simulate waiting"}
            </Button>
          </div>
          <p>
            Preview submissions:{" "}
            <output aria-label="Preview submissions">{submitted}</output>
          </p>
        </form>
      </Surface>
      <Surface aria-labelledby="gallery-learning">
        <LearningIntro
          id="gallery-learning"
          title="Learning content"
          description="A practical meaning with a clear next step."
        />
        <p className="mt-4 break-words font-[family-name:var(--vn-headword-font)] text-4xl font-semibold">
          invite
        </p>
        <p className="mt-2 italic text-neutral-600">verb</p>
        <p className="mt-3 text-lg">
          To ask someone to come to an event or do something with you.
        </p>
        <p className="mt-3">“I want to invite my friends to dinner.”</p>
        <LearningNote>
          <p>Try writing about someone you would like to invite.</p>
        </LearningNote>
      </Surface>
      <Surface tone="secondary" aria-labelledby="gallery-feedback">
        <h2 id="gallery-feedback" className="text-xl font-semibold">
          Feedback and recovery
        </h2>
        <div className="mt-4 space-y-4">
          <StatusMessage tone="info" role="note">
            No words are due in this synthetic example. Discover a useful word
            when you are ready.
          </StatusMessage>
          <StatusMessage tone="success" role="note">
            The preview action is complete. This is an example, not a saved
            learning record.
          </StatusMessage>
          <StatusMessage tone="error" role="note">
            We could not load this example. Your sentence is still here.
          </StatusMessage>
          <Button
            variant="secondary"
            onClick={() => setMessage("Retry completed in the preview.")}
          >
            Retry preview
          </Button>
        </div>
      </Surface>
    </div>
  );
}
