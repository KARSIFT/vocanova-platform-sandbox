import { ActionLink } from "@/ui/controls";
import { PageContainer, PageHeader } from "@/ui/surface";

export default function SavedWordNotFound() {
  return (
    <PageContainer>
      <PageHeader
        title="Saved word not found"
        lede="It may have been removed from your vocabulary."
      />
      <ActionLink variant="secondary" href="/words">
        Back to saved vocabulary
      </ActionLink>
    </PageContainer>
  );
}
