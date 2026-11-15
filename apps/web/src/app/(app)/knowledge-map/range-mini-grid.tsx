import { useId } from "react";
import type { KnowledgeMapData } from "./models";
import { rangePreviewPaths } from "./range-preview-model";
import styles from "./workspace.module.css";

/** Self-contained, decorative grid: the enclosing control provides readable counts. */
export function RangeMiniGrid({
  band,
}: {
  band: KnowledgeMapData["bands"][number];
}) {
  const patternId = `word-range-${useId().replaceAll(":", "")}`;
  const paths = rangePreviewPaths(band);
  const knownHeight = (band.knownCount / band.totalWords) * 100;
  const savedHeight = (band.learningCount / band.totalWords) * 100;
  return (
    <svg
      viewBox="0 0 100 100"
      className={styles.atlasGrid}
      aria-hidden="true"
      focusable="false"
    >
      {paths ? (
        <>
          <defs>
            <pattern
              id={patternId}
              width="10"
              height="10"
              patternUnits="userSpaceOnUse"
            >
              <rect
                x="1"
                y="1"
                width="8"
                height="8"
                className={styles.atlasUnseen}
              />
            </pattern>
          </defs>
          <rect width="100" height="100" fill={`url(#${patternId})`} />
          {paths.k && <path data-state="known" d={paths.k} />}
          {paths.l && <path data-state="learning" d={paths.l} />}
          {paths.m && <path data-state="learned" d={paths.m} />}
        </>
      ) : (
        <>
          <rect width="100" height="100" fill="var(--color-white)" />
          <rect
            y={100 - knownHeight - savedHeight}
            width="100"
            height={savedHeight}
            data-state="learning"
          />
          <rect
            y={100 - knownHeight}
            width="100"
            height={knownHeight}
            data-state="known"
          />
        </>
      )}
    </svg>
  );
}
