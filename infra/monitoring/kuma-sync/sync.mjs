import { validateMonitoringInventory } from "../validate-inventory.mjs";
import { applyOperations, SyncApplyError } from "./apply.mjs";
import { planSyncOperations } from "./plan.mjs";
import { createRedactingLogger, redactSecrets } from "./redact.mjs";

export class SyncValidationError extends Error {
  constructor(errors) {
    super(errors.join("; "));
    this.name = "SyncValidationError";
    this.errors = errors;
  }
}

export async function syncKumaMonitors({
  monitorsDocument,
  syntheticsDocument,
  client,
  logger = console,
  requireCanonicalIds = true,
}) {
  const redactingLogger = createRedactingLogger(logger);

  const inventoryErrors = validateMonitoringInventory({
    monitorsDocument,
    syntheticsDocument,
    requireCanonicalIds,
  });

  if (inventoryErrors.length > 0) {
    throw new SyncValidationError(inventoryErrors);
  }

  const ownershipMarker = monitorsDocument.kuma.ownership_marker;
  const inventoryMonitors = monitorsDocument.availability_monitors;

  await client.connect();
  const remoteMonitors = await client.listMonitors();

  const { operations, errors: planErrors } = planSyncOperations({
    inventoryMonitors,
    remoteMonitors,
    ownershipMarker,
  });

  if (planErrors.length > 0) {
    throw new SyncValidationError(planErrors);
  }

  if (operations.length === 0) {
    redactingLogger.info("Kuma inventory already matches repository inventory (no-op).");
    return { changed: 0, applied: [] };
  }

  const { applied, changed } = await applyOperations({
    client,
    operations,
    logger: redactingLogger,
  });

  return { changed, applied };
}

export function formatSyncFailure(error) {
  const raw =
    error instanceof SyncValidationError || error instanceof SyncApplyError
      ? error.message
      : String(error?.message ?? error);
  return redactSecrets(raw);
}
