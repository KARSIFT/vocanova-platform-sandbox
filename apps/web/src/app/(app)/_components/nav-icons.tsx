/** Shared primary-destination icons; the visible label always carries the name. */
const iconClass =
  "size-[1.15rem] shrink-0 fill-none stroke-current stroke-[1.9]";

export function HomeIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className={iconClass}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="m4 10.5 8-6.25 8 6.25v8.25a1 1 0 0 1-1 1h-4.25v-5.5h-5.5v5.5H5a1 1 0 0 1-1-1V10.5Z"
      />
    </svg>
  );
}

export function JourneyIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className={iconClass}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M9 4.5 4.5 6.25v13.25L9 17.75l6 1.75 4.5-1.75V4.5L15 6.25 9 4.5Zm0 0v13.25m6-11.5v13.25"
      />
    </svg>
  );
}

export function ProgressIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className={iconClass}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M5 19.25V12m7 7.25V5m7 14.25v-9"
      />
    </svg>
  );
}
