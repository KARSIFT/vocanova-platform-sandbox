"use client";

import { useEffect, useId, useRef, useState, type RefObject } from "react";
import type { DailyLessonAction } from "@vocanova/api-client";

type ReviewChoice = NonNullable<DailyLessonAction["nextReview"]>;
const intervals: { value: ReviewChoice; label: string }[] = [
  { value: "1d", label: "In 1 day" },
  { value: "3d", label: "In 3 days" },
  { value: "7d", label: "In 7 days" },
  { value: "14d", label: "In 14 days" },
  { value: "30d", label: "In 30 days" },
  { value: "60d", label: "In 60 days" },
  { value: "120d", label: "In 120 days" },
];

/** A local schedule choice. Only Continue sends it to the server. */
export function ReviewSchedule({
  value,
  automaticLabel,
  disabled,
  onChange,
  nextAction,
}: {
  value: ReviewChoice;
  automaticLabel: string;
  disabled: boolean;
  onChange: (value: ReviewChoice) => void;
  nextAction: RefObject<HTMLButtonElement | null>;
}) {
  const [open, setOpen] = useState(false);
  const [focused, setFocused] = useState(0);
  const container = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const options = useRef<(HTMLButtonElement | null)[]>([]);
  const menuId = useId();
  const choices = [
    { value: "automatic" as const, label: automaticLabel },
    ...intervals,
    { value: "known" as const, label: "I know this word" },
  ];
  const selected = Math.max(
    0,
    choices.findIndex((choice) => choice.value === value),
  );
  const visibleValue =
    value === "automatic"
      ? automaticLabel.replace(/^Automatic · /u, "")
      : choices[selected]!.label;

  useEffect(() => {
    if (disabled) setOpen(false);
  }, [disabled]);

  useEffect(() => {
    if (open && !disabled) options.current[focused]?.focus();
  }, [open, focused, disabled]);

  useEffect(() => {
    if (!open) return;
    function dismiss(event: PointerEvent) {
      if (
        event.target instanceof Node &&
        !container.current?.contains(event.target)
      )
        setOpen(false);
    }
    document.addEventListener("pointerdown", dismiss);
    return () => document.removeEventListener("pointerdown", dismiss);
  }, [open]);

  function close(returnFocus = false) {
    setOpen(false);
    if (returnFocus) trigger.current?.focus();
  }
  function show() {
    setFocused(selected);
    setOpen(true);
  }

  return (
    <div
      ref={container}
      className="review-schedule"
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) close();
      }}
    >
      <button
        ref={trigger}
        type="button"
        className="review-schedule-trigger"
        aria-label={`Show this word again: ${choices[selected]!.label}`}
        aria-haspopup="menu"
        aria-expanded={open && !disabled}
        aria-controls={open && !disabled ? menuId : undefined}
        disabled={disabled}
        onClick={() => (open ? close() : show())}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            show();
          } else if (event.key === "Escape" && open) {
            event.preventDefault();
            close(true);
          }
        }}
      >
        <span className="review-schedule-caption">Next review</span>
        <span className="review-schedule-value">{visibleValue}</span>
        <svg
          aria-hidden="true"
          viewBox="0 0 20 20"
          className="review-schedule-chevron"
        >
          <path d="m5 8 5 5 5-5" />
        </svg>
      </button>
      <span className="sr-only" aria-live="polite" aria-atomic="true">
        {`Next review: ${choices[selected]!.label}`}
      </span>
      {open && !disabled && (
        <div
          id={menuId}
          role="menu"
          aria-label="Next review"
          className="review-schedule-menu"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              close(true);
            } else if (event.key === "Tab") {
              event.preventDefault();
              close();
              if (event.shiftKey) trigger.current?.focus();
              else (nextAction.current ?? trigger.current)?.focus();
            } else if (
              [
                "ArrowDown",
                "ArrowRight",
                "ArrowUp",
                "ArrowLeft",
                "Home",
                "End",
              ].includes(event.key)
            ) {
              event.preventDefault();
              const next =
                event.key === "Home"
                  ? 0
                  : event.key === "End"
                    ? choices.length - 1
                    : (focused +
                        (["ArrowDown", "ArrowRight"].includes(event.key)
                          ? 1
                          : choices.length - 1)) %
                      choices.length;
              setFocused(next);
            }
          }}
        >
          {choices.map((choice, index) => (
            <button
              key={choice.value}
              ref={(element) => {
                options.current[index] = element;
              }}
              type="button"
              role="menuitemradio"
              aria-checked={value === choice.value}
              tabIndex={index === focused ? 0 : -1}
              className={`review-schedule-option${index === 0 || choice.value === "known" ? " review-schedule-option-wide" : ""}`}
              onFocus={() => setFocused(index)}
              onClick={() => {
                onChange(choice.value);
                close(true);
              }}
            >
              <span>{choice.label}</span>
              <svg
                aria-hidden="true"
                viewBox="0 0 20 20"
                className="review-schedule-check"
              >
                <path d="m4 10 4 4 8-8" />
              </svg>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
