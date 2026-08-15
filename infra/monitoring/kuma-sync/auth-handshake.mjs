/**
 * Arm a one-shot monitorList waiter before login so a post-auth push that
 * arrives before (or with) the login ack is not missed.
 */
export function createMonitorListWaiter(socket, timeoutMs = 10000) {
  let settled = false;
  let onList;
  let timer;
  let rejectPromise;

  const promise = new Promise((resolve, reject) => {
    rejectPromise = reject;

    timer = setTimeout(() => {
      if (settled) {
        return;
      }
      settled = true;
      if (onList) {
        socket.off("monitorList", onList);
      }
      reject(new Error("timed out waiting for monitorList after login"));
    }, timeoutMs);

    onList = (monitorList) => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timer);
      resolve(monitorList ?? {});
    };
    socket.once("monitorList", onList);
  });

  // Prevent unhandledRejection when cancel() runs before the caller awaits.
  promise.catch(() => {});

  return {
    promise,
    cancel(reason = new Error("monitorList wait cancelled")) {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timer);
      if (onList) {
        socket.off("monitorList", onList);
      }
      rejectPromise(reason);
    },
  };
}

/**
 * Authenticate and capture the initial monitor list. The monitorList listener
 * is registered before the login emit so Kuma's common post-login push cannot
 * race ahead of the waiter.
 */
export async function authenticateAndLoadMonitors(
  socket,
  { username, password, timeoutMs = 10000 },
) {
  const listWait = createMonitorListWaiter(socket, timeoutMs);

  try {
    const loginResult = await new Promise((resolve) => {
      socket.emit("login", { username, password }, (response) => {
        resolve(response ?? { ok: false, msg: "empty login response" });
      });
    });

    if (!loginResult.ok) {
      listWait.cancel(new Error("login failed; abandoning monitorList wait"));
      throw new Error(loginResult.msg ?? "Kuma login failed");
    }

    return await listWait.promise;
  } catch (error) {
    listWait.cancel(error);
    throw error;
  }
}
