import { KnowledgeMapView } from "./map-view";

export default async function KnowledgeMapPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  return <KnowledgeMapView searchParams={searchParams} />;
}
