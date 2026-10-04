import { PageContainer, Surface } from "@/ui/surface";
export default function StoriesLoading() {
  return (
    <PageContainer className="max-w-[44rem]">
      <Surface>
        <p role="status" className="text-neutral-700">
          Loading your story…
        </p>
      </Surface>
    </PageContainer>
  );
}
