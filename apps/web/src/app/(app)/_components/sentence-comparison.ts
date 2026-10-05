export interface SentenceChange {
  text: string;
  changed: boolean;
}

/** Exact, case-sensitive token comparison. This describes edits, not their reason. */
export function compareSentences(
  original: string,
  suggested: string,
): {
  original: SentenceChange[];
  suggested: SentenceChange[];
  hasChanges: boolean;
} {
  const tokens = (text: string) =>
    text.match(/[\p{L}\p{N}\p{M}]+|\s+|[^\p{L}\p{N}\p{M}\s]/gu) ?? [];
  const before = tokens(original);
  const after = tokens(suggested);
  // Provider output is not constrained like the 300-character learner input.
  // Bound the comparison work; unusually long text still renders verbatim.
  if (before.length * after.length > 100_000) {
    const changed = original !== suggested;
    return {
      original: [{ text: original, changed }],
      suggested: [{ text: suggested, changed }],
      hasChanges: changed,
    };
  }
  const lengths = Array.from(
    { length: before.length + 1 },
    () => new Uint32Array(after.length + 1),
  );
  for (let i = before.length - 1; i >= 0; i -= 1) {
    for (let j = after.length - 1; j >= 0; j -= 1) {
      lengths[i]![j] =
        before[i] === after[j]
          ? lengths[i + 1]![j + 1]! + 1
          : Math.max(lengths[i + 1]![j]!, lengths[i]![j + 1]!);
    }
  }
  const left: SentenceChange[] = [];
  const right: SentenceChange[] = [];
  const append = (parts: SentenceChange[], text: string, changed: boolean) => {
    const last = parts.at(-1);
    if (last && last.changed === changed) last.text += text;
    else parts.push({ text, changed });
  };
  let i = 0;
  let j = 0;
  while (i < before.length || j < after.length) {
    if (i < before.length && j < after.length && before[i] === after[j]) {
      append(left, before[i++]!, false);
      append(right, after[j++]!, false);
    } else if (
      i < before.length &&
      (j === after.length || lengths[i + 1]![j]! >= lengths[i]![j + 1]!)
    ) {
      append(left, before[i++]!, true);
    } else {
      append(right, after[j++]!, true);
    }
  }
  return {
    original: left,
    suggested: right,
    hasChanges: original !== suggested,
  };
}
