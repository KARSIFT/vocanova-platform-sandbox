export function createMockKumaClient(initialMonitors = {}, options = {}) {
  const monitors = structuredClone(initialMonitors);
  let nextId =
    options.nextId ??
    Math.max(0, ...Object.keys(monitors).map((id) => Number(id))) + 1;

  const calls = [];
  let failAfterApplied = options.failAfterApplied ?? null;
  let failDeleteMonitor = options.failDeleteMonitor ?? false;
  let appliedMutations = 0;

  function shouldFailApply(operation) {
    if (!failAfterApplied) {
      return false;
    }
    if (failAfterApplied.operation && failAfterApplied.operation !== operation) {
      return false;
    }
    return appliedMutations >= failAfterApplied.after;
  }

  function recordCall(call) {
    calls.push(call);
    if (call.op === "add" || call.op === "editMonitor") {
      if (shouldFailApply(call.op)) {
        throw new Error(
          failAfterApplied.message ?? `simulated ${call.op} failure`,
        );
      }
      appliedMutations += 1;
    }

    if (call.op === "deleteMonitor" && failDeleteMonitor) {
      throw new Error("simulated deleteMonitor rollback failure");
    }
  }

  return {
    calls,
    monitors,
    async connect() {},
    async disconnect() {},
    async listMonitors() {
      return structuredClone(monitors);
    },
    async add(monitor) {
      recordCall({ op: "add", monitor: structuredClone(monitor) });
      const monitorID = nextId++;
      monitors[monitorID] = {
        ...structuredClone(monitor),
        id: monitorID,
      };
      return { ok: true, monitorID };
    },
    async editMonitor(monitor) {
      recordCall({ op: "editMonitor", monitor: structuredClone(monitor) });
      const monitorID = Number(monitor.id);
      if (!monitors[monitorID]) {
        return { ok: false, msg: `monitor ${monitorID} not found` };
      }
      monitors[monitorID] = {
        ...monitors[monitorID],
        ...structuredClone(monitor),
        id: monitorID,
      };
      return { ok: true, monitorID };
    },
    async deleteMonitor(monitorID) {
      recordCall({ op: "deleteMonitor", monitorID: Number(monitorID) });
      const id = Number(monitorID);
      if (!monitors[id]) {
        return { ok: false, msg: `monitor ${id} not found` };
      }
      delete monitors[id];
      return { ok: true };
    },
    setFailAfterApplied(config) {
      failAfterApplied = config;
    },
    setFailDeleteMonitor(value = true) {
      failDeleteMonitor = value;
    },
  };
}
