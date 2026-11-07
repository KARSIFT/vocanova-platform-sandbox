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

/** A centred eight-tooth cog, matching the shared navigation stroke. */
export function SettingsIcon() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className={iconClass}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M18.65 9.24 L19.02 10.38 L21.28 10.53 L21.28 13.47 L19.02 13.62 L18.65 14.76 L18.11 15.82 L19.60 17.53 L17.53 19.60 L15.82 18.11 L14.76 18.65 L13.62 19.02 L13.47 21.28 L10.53 21.28 L10.38 19.02 L9.24 18.65 L8.18 18.11 L6.47 19.60 L4.40 17.53 L5.89 15.82 L5.35 14.76 L4.98 13.62 L2.72 13.47 L2.72 10.53 L4.98 10.38 L5.35 9.24 L5.89 8.18 L4.40 6.47 L6.47 4.40 L8.18 5.89 L9.24 5.35 L10.38 4.98 L10.53 2.72 L13.47 2.72 L13.62 4.98 L14.76 5.35 L15.82 5.89 L17.53 4.40 L19.60 6.47 L18.11 8.18Z"
      />
      <circle cx="12" cy="12" r="3.1" />
    </svg>
  );
}
