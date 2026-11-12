export const DISPLAY_NAME_MAX_LENGTH = 80;

/** Match the API's Unicode code-point limit rather than UTF-16 input units. */
export function isDisplayNameWithinLimit(value: string): boolean {
  return Array.from(value).length <= DISPLAY_NAME_MAX_LENGTH;
}
