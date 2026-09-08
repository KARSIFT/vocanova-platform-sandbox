export type ReviewSeedAction = "remove" | "save";

/**
 * Returns the real saved-word transitions the persistent staging account must
 * perform before opening Review. Restoring an existing saved word resets its
 * schedule to immediately due; a new word only needs the ordinary save.
 */
export function reviewSeedActions(initiallySaved: boolean): ReviewSeedAction[] {
  return initiallySaved ? ["remove", "save"] : ["save"];
}
