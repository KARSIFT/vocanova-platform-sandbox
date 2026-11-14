import { StatusMessage } from "@/ui/status-message";
import { PageContainer } from "@/ui/surface";
export default function WritingLoading() {
  return (
    <PageContainer>
      <StatusMessage as="p" role="status" className="text-neutral-700">
        Loading writing topics…
      </StatusMessage>
    </PageContainer>
  );
}
