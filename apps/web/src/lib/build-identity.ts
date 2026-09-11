export interface BuildIdentity {
  builtAt: string;
  commit: string;
  environment: "development" | "production" | "staging" | "test" | "unknown";
  version: string;
}

interface BuildIdentityInput {
  builtAt?: string;
  commit?: string;
  environment?: string;
  version?: string;
}

const UNKNOWN = "unknown";
const VERSION_PATTERN = /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/;
const COMMIT_PATTERN = /^[0-9a-f]{40}$/;
const ISO_TIMESTAMP_PATTERN =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})$/;

function sanitizeVersion(value?: string): string {
  const trimmed = value?.trim();
  return trimmed && trimmed.length <= 64 && VERSION_PATTERN.test(trimmed)
    ? trimmed
    : UNKNOWN;
}

function sanitizeCommit(value?: string): string {
  const trimmed = value?.trim();
  return trimmed && COMMIT_PATTERN.test(trimmed) ? trimmed : UNKNOWN;
}

function sanitizeEnvironment(value?: string): BuildIdentity["environment"] {
  return value === "development" ||
    value === "test" ||
    value === "staging" ||
    value === "production"
    ? value
    : UNKNOWN;
}

function sanitizeBuiltAt(value?: string): string {
  const trimmed = value?.trim();
  if (!trimmed || !ISO_TIMESTAMP_PATTERN.test(trimmed)) {
    return UNKNOWN;
  }
  const timestamp = Date.parse(trimmed);
  return Number.isFinite(timestamp)
    ? new Date(timestamp).toISOString()
    : UNKNOWN;
}

export function getBuildIdentity(input: BuildIdentityInput): BuildIdentity {
  return {
    version: sanitizeVersion(input.version),
    commit: sanitizeCommit(input.commit),
    environment: sanitizeEnvironment(input.environment),
    builtAt: sanitizeBuiltAt(input.builtAt),
  };
}

// These public values are deliberately injected before `next build`. Direct
// property access lets Next embed them in the browser bundle; there is no
// runtime configuration or request data in this identity.
export const buildIdentity = getBuildIdentity({
  version: process.env.NEXT_PUBLIC_VOCANOVA_BUILD_VERSION,
  commit: process.env.NEXT_PUBLIC_VOCANOVA_BUILD_COMMIT,
  environment: process.env.NEXT_PUBLIC_VOCANOVA_BUILD_ENVIRONMENT,
  builtAt: process.env.NEXT_PUBLIC_VOCANOVA_BUILD_BUILT_AT,
});
