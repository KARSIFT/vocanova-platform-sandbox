import Link from "@/ui/link";
import type { Ref } from "react";
import {
  earlierDiscoveryBand,
  journeyRangeHref,
  type KnowledgeMapData,
} from "../knowledge-map/models";
import styles from "../knowledge-map/workspace.module.css";

export function DiscoveryFinished({
  data,
  active,
  headingRef,
}: {
  data: KnowledgeMapData;
  active: boolean;
  headingRef?: Ref<HTMLHeadingElement>;
}) {
  const earlier = earlierDiscoveryBand(data);
  return (
    <>
      <h1 className={styles.title} ref={headingRef} tabIndex={-1}>
        {earlier
          ? "You’ve finished this range"
          : "You’ve explored the whole map"}
      </h1>
      <p className={styles.definition}>
        {earlier
          ? "Earlier ranges still have new words."
          : active
            ? "Keep the words you are learning fresh with a short review."
            : "Choose a word on the map to revisit it."}
      </p>
      <div className={styles.startActions}>
        {earlier ? (
          <Link
            className={styles.primary}
            href={`/discovery?band=${earlier.number}&active=1`}
          >
            Continue with earlier words
          </Link>
        ) : active ? (
          <Link className={styles.primary} href="/reviews">
            Review words
          </Link>
        ) : null}
        <Link
          className={earlier || active ? styles.secondary : styles.primary}
          href={journeyRangeHref(data.currentBand.number)}
        >
          {active ? "View map" : "Open your word map"}
        </Link>
      </div>
    </>
  );
}
