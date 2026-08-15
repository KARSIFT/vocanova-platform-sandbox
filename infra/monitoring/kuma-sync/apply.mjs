import { buildEditPayload } from "./monitor-payload.mjs";
import { redactSecrets } from "./redact.mjs";

export class SyncApplyError extends Error {
  constructor(message, { applied = [], rollbackErrors = [] } = {}) {
    super(redactSecrets(message));
    this.name = "SyncApplyError";
    this.applied = applied;
    this.rollbackErrors = rollbackErrors;
  }
}

async function rollbackApplied({ client, applied, rollbackFailures }) {
  for (const record of [...applied].reverse()) {
    try {
      if (record.type === "create") {
        const result = await client.deleteMonitor(record.kumaId);
        if (!result.ok) {
          rollbackFailures.push(
            `rollback deleteMonitor(${record.kumaId}) failed: ${result.msg ?? "unknown error"}`,
          );
        }
      } else {
        const payload = buildEditPayload(record.previous, record.kumaId);
        const result = await client.editMonitor(payload);
        if (!result.ok) {
          rollbackFailures.push(
            `rollback editMonitor(${record.kumaId}) failed: ${result.msg ?? "unknown error"}`,
          );
        }
      }
    } catch (error) {
      rollbackFailures.push(
        `rollback ${record.type}(${record.kumaId}) threw: ${error.message}`,
      );
    }
  }
}

export async function applyOperations({ client, operations, logger }) {
  const applied = [];

  try {
    for (const operation of operations) {
      if (operation.type === "create") {
        const result = await client.add(operation.desired);
        if (!result.ok) {
          throw new SyncApplyError(
            `add failed for ${operation.inventoryId}: ${result.msg ?? "unknown error"}`,
            { applied },
          );
        }

        applied.push({
          type: "create",
          inventoryId: operation.inventoryId,
          kumaId: Number(result.monitorID),
          previous: null,
        });
        logger.info(
          `created monitor ${operation.inventoryId} (kuma id ${result.monitorID})`,
        );
        continue;
      }

      const payload = buildEditPayload(operation.desired, operation.kumaId);
      const result = await client.editMonitor(payload);
      if (!result.ok) {
        throw new SyncApplyError(
          `editMonitor failed for ${operation.inventoryId}: ${result.msg ?? "unknown error"}`,
          { applied },
        );
      }

      applied.push({
        type: "update",
        inventoryId: operation.inventoryId,
        kumaId: operation.kumaId,
        previous: operation.previous,
      });
      logger.info(`updated monitor ${operation.inventoryId} (kuma id ${operation.kumaId})`);
    }

    return { applied, changed: applied.length };
  } catch (error) {
    const rollbackFailures = [];
    await rollbackApplied({ client, applied, rollbackFailures });

    if (rollbackFailures.length > 0) {
      throw new SyncApplyError(
        `apply failed and rollback was incomplete: ${rollbackFailures.join("; ")}`,
        { applied, rollbackErrors: rollbackFailures },
      );
    }

    if (error instanceof SyncApplyError) {
      throw new SyncApplyError(error.message, {
        applied,
        rollbackErrors: rollbackFailures,
      });
    }

    throw new SyncApplyError(error.message, { applied, rollbackErrors: rollbackFailures });
  }
}
