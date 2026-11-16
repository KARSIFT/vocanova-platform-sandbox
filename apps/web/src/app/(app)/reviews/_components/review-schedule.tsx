"use client";

import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type RefObject,
} from "react";
import type { DailyLessonAction } from "@vocanova/api-client";

import { draftReviewDate, relativeReviewDate } from "./daily-review-view";

type ReviewChoice = NonNullable<DailyLessonAction["nextReview"]>;
const intervals = ["1d", "3d", "7d", "14d", "30d", "60d", "120d"] as const;

/** A local schedule choice. Only Continue sends it to the server. */
export function ReviewSchedule({
  value,
  automaticLabel,
  automaticExact,
  automaticAt,
  timeZone,
  disabled,
  onChange,
  nextAction,
}: {
  value: ReviewChoice;
  automaticLabel: string;
  automaticExact?: string;
  automaticAt?: string;
  timeZone?: string;
  disabled: boolean;
  onChange: (value: ReviewChoice) => void;
  nextAction: RefObject<HTMLButtonElement | null>;
}) {
  const [now, setNow] = useState(() => new Date());
  const [open, setOpen] = useState(false);
  const [focused, setFocused] = useState(0);
  const [menuLayout, setMenuLayout] = useState<{
    height: number;
    gap: number;
    compact: boolean;
  }>();
  const container = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const options = useRef<(HTMLButtonElement | null)[]>([]);
  const menuId = useId();
  const intervalChoices = intervals.map((interval) => ({
    value: interval,
    // Legacy responses without a zone can promise elapsed time, not a date.
    ...(draftReviewDate(interval, timeZone, now) ?? {
      label: `In ${Number(interval.slice(0, -1)) * 24} hours`,
      exact: undefined,
    }),
  }));
  const automaticDate = automaticAt
    ? relativeReviewDate(automaticAt, timeZone, now)
    : undefined;
  const automaticUTC = automaticAt
    ? relativeReviewDate(automaticAt, "UTC", now)
    : undefined;
  const automaticChoice = {
    value: "automatic" as const,
    ...(automaticDate ??
      (automaticUTC
        ? { label: automaticUTC.exact, exact: undefined }
        : {
            label:
              automaticLabel.replace(/^Automatic(?: · )?/u, "") ||
              "Scheduled review",
            exact: automaticExact,
          })),
  };
  // The checked row represents a calendar date. Keep the original server choice
  // and its exact time until the learner explicitly picks an elapsed interval.
  let automaticInterval: (typeof intervals)[number] | undefined;
  if (automaticAt && timeZone && automaticDate) {
    const calendar = new Intl.DateTimeFormat("en", {
      timeZone,
      year: "numeric",
      month: "numeric",
      day: "numeric",
    });
    const scheduledDay = calendar.format(new Date(automaticAt));
    automaticInterval = intervals.find(
      (interval) =>
        calendar.format(
          new Date(now.getTime() + Number(interval.slice(0, -1)) * 86_400_000),
        ) === scheduledDay,
    );
  }
  const choices = [
    ...(automaticInterval ? [] : [automaticChoice]),
    ...intervalChoices,
    { value: "known" as const, label: "I know this word", exact: undefined },
  ];
  const selectedValue =
    value === "automatic" ? (automaticInterval ?? "automatic") : value;
  const selected = Math.max(
    0,
    choices.findIndex((choice) => choice.value === selectedValue),
  );
  const selectedChoice =
    value === "automatic" ? automaticChoice : choices[selected]!;

  const measureMenu = useCallback(() => {
    if (!trigger.current) return;
    const rootFont = parseFloat(
      getComputedStyle(document.documentElement).fontSize,
    );
    const gap = Math.max(8, Math.min(24, rootFont * 0.75));
    const headerBottom =
      document.querySelector("header")?.getBoundingClientRect().bottom ?? 0;
    const height = Math.floor(
      Math.max(
        0,
        trigger.current.getBoundingClientRect().top -
          gap -
          Math.max(0, headerBottom) -
          8,
      ),
    );
    const compact = height < rootFont * 6;
    setMenuLayout((current) =>
      current?.height === height &&
      current.gap === gap &&
      current.compact === compact
        ? current
        : { height, gap, compact },
    );
  }, []);

  useEffect(() => {
    if (!open) return;
    const header = document.querySelector("header");
    const observer = new ResizeObserver(measureMenu);
    if (trigger.current) observer.observe(trigger.current);
    if (header) observer.observe(header);
    window.addEventListener("resize", measureMenu);
    window.addEventListener("scroll", measureMenu, { passive: true });
    window.visualViewport?.addEventListener("resize", measureMenu);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", measureMenu);
      window.removeEventListener("scroll", measureMenu);
      window.visualViewport?.removeEventListener("resize", measureMenu);
    };
  }, [open, measureMenu]);

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
    measureMenu();
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
        aria-label={`Show this word again: ${selectedChoice.label}${selectedChoice.exact ? `. ${selectedChoice.exact}` : ""}`}
        title={selectedChoice.exact}
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
        <span className="review-schedule-value">{selectedChoice.label}</span>
        <svg
          aria-hidden="true"
          viewBox="0 0 20 20"
          className="review-schedule-chevron"
        >
          <path d="m5 8 5 5 5-5" />
        </svg>
      </button>
      <span className="sr-only">{`Next review: ${selectedChoice.label}`}</span>
      {open && !disabled && (
        <div
          id={menuId}
          role="menu"
          aria-label="Next review"
          className="review-schedule-menu"
          data-compact={menuLayout?.compact || undefined}
          style={{
            maxHeight: menuLayout?.height,
            bottom: menuLayout ? `calc(100% + ${menuLayout.gap}px)` : undefined,
          }}
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
          {choices.map((choice, index) => {
            const displayedChoice =
              value === "automatic" && index === selected
                ? automaticChoice
                : choice;
            return (
              <button
                key={choice.value}
                ref={(element) => {
                  options.current[index] = element;
                }}
                type="button"
                role="menuitemradio"
                aria-checked={index === selected}
                tabIndex={index === focused ? 0 : -1}
                className={`review-schedule-option${choice.value === "automatic" || choice.value === "known" ? " review-schedule-option-wide" : ""}`}
                onFocus={() => setFocused(index)}
                onClick={() => {
                  onChange(
                    value === "automatic" && choice.value === automaticInterval
                      ? "automatic"
                      : choice.value,
                  );
                  close(true);
                }}
              >
                <span>
                  {displayedChoice.label}
                  {displayedChoice.exact && (
                    <span className="sr-only">. {displayedChoice.exact}</span>
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
            );
          })}
        </div>
      )}
    </div>
  );
}
