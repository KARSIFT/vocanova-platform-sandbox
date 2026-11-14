import { StatusMessage } from "@/ui/status-message";
import { PageContainer, Surface } from "@/ui/surface";
export default function StoriesLoading() {
  return (
    <PageContainer className="max-w-[44rem]">
      <Surface>
        <StatusMessage as="p" role="status" className="text-neutral-700">
          Loading your story…
        </StatusMessage>
      </Surface>
    </PageContainer>
  );
}
