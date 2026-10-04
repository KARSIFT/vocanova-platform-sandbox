"use client";

import Link from "next/link";
import { PageContainer, Surface } from "@/ui/surface";
import {
  primaryAction,
  textLink,
} from "../practice/_components/practice-styles";
export default function StoryError({ reset }: { reset: () => void }) {
  return (
    <PageContainer className="max-w-[44rem]">
      <Surface>
        <h1 className="text-2xl font-bold text-neutral-900">
          Your story could not load
        </h1>
        <p role="alert" className="my-4 text-neutral-700">
          Try again to load your saved progress.
        </p>
        <button type="button" className={primaryAction} onClick={reset}>
          Try again
        </button>
        <Link href="/stories" className={`${textLink} ml-4`}>
          Back to Stories
        </Link>
      </Surface>
    </PageContainer>
  );
}
