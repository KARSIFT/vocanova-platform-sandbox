"use client";

import { useEffect, useId, useRef, useState, type RefObject } from "react";
import type { DailyLessonAction } from "@vocanova/api-client";

import { draftReviewDate } from "./daily-review-view";

type ReviewChoice = NonNullable<DailyLessonAction["nextReview"]>;
const intervals = ["1d", "3d", "7d", "14d", "30d", "60d", "120d"] as const;

/** A local schedule choice. Only Continue sends it to the server. */
export function ReviewSchedule({
  value,
  automaticLabel,
  automaticExact,
  timeZone,
  disabled,
  onChange,
  nextAction,
}: {
  value: ReviewChoice;
  automaticLabel: string;
  automaticExact?: string;
  timeZone?: string;
  disabled: boolean;
  onChange: (value: ReviewChoice) => void;
  nextAction: RefObject<HTMLButtonElement | null>;
}) {
  const [now, setNow] = useState(() => new Date());
  const [open, setOpen] = useState(false);
  const [focused, setFocused] = useState(0);
  const container = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const options = useRef<(HTMLButtonElement | null)[]>([]);
  const menuId = useId();
  const choices = [
    {
      value: "automatic" as const,
      label: automaticLabel,
      exact: automaticExact,
    },
    ...intervals.map((interval) => ({
      value: interval,
      // Legacy responses without a zone can promise elapsed time, not a date.
      ...(draftReviewDate(interval, timeZone, now) ?? {
        label: `In ${Number(interval.slice(0, -1)) * 24} hours`,
        exact: undefined,
      }),
    })),
    { value: "known" as const, label: "I know this word", exact: undefined },
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
    const timer = window.setInterval(() => setNow(new Date()), 30_000);
    return () => window.clearInterval(timer);
  }, []);

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
    setNow(new Date());
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
        aria-label={`Show this word again: ${choices[selected]!.label}${choices[selected]!.exact ? `. ${choices[selected]!.exact}` : ""}`}
        title={choices[selected]!.exact}
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
      <span className="sr-only">
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
              <span>
                {choice.label}
                {choice.exact && (
                  <span className="sr-only">. {choice.exact}</span>
                )}
              </span>
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
