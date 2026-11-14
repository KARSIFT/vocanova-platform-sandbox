import { notFound } from "next/navigation";
import { PageContainer } from "@/ui/surface";
import { LearningIntro } from "@/ui/learning-guide";
import { ComponentExamples } from "./component-examples";

export const dynamic = "force-dynamic";
export const metadata = {
  title: "Component gallery · Vocanova",
  robots: { index: false, follow: false },
};

export default function ComponentGalleryPage() {
  if (process.env.VOCANOVA_COMPONENT_GALLERY !== "true") notFound();
  return (
    <main id="main-content">
      <PageContainer>
        <LearningIntro
          headingLevel={1}
          title="Vocanova component gallery"
          description="Development preview with synthetic examples. These controls do not save learning progress or contact a feedback provider."
        />
        <ComponentExamples />
      </PageContainer>
    </main>
  );
}
