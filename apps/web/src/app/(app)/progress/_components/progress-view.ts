// Progress displays a bounded recent-vocabulary sample. Keep this copy
// independent from the returned page length: that length is not the learner's
// total saved-word count when more results exist beyond the requested page.
export function getSavedVocabularySummary(limit: number): string {
  return `Showing up to ${limit} recently saved words.`;
}
