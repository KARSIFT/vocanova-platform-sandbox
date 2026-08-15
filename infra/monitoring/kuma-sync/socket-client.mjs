import { io } from "socket.io-client";

function waitForMonitorList(socket, timeoutMs = 10000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("timed out waiting for monitorList after login"));
    }, timeoutMs);

    socket.once("monitorList", (monitorList) => {
      clearTimeout(timer);
      resolve(monitorList ?? {});
    });
  });
}

export async function createSocketKumaClient({
  baseUrl,
  username,
  password,
  timeoutMs = 10000,
}) {
  if (!baseUrl || !username || !password) {
    throw new Error("baseUrl, username, and password are required for Kuma login");
  }

  const socket = io(baseUrl, {
    transports: ["polling", "websocket"],
    reconnection: false,
    timeout: timeoutMs,
  });

  await new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("timed out connecting to Kuma Socket.IO endpoint"));
    }, timeoutMs);

    socket.once("connect", () => {
      clearTimeout(timer);
      resolve();
    });
    socket.once("connect_error", (error) => {
      clearTimeout(timer);
      reject(error);
    });
  });

  const loginResult = await new Promise((resolve) => {
    socket.emit("login", { username, password }, (response) => {
      resolve(response ?? { ok: false, msg: "empty login response" });
    });
  });

  if (!loginResult.ok) {
    socket.disconnect();
    throw new Error(loginResult.msg ?? "Kuma login failed");
  }

  const initialMonitors = await waitForMonitorList(socket, timeoutMs);

  return {
    socket,
    async connect() {},
    async disconnect() {
      socket.disconnect();
    },
    async listMonitors() {
      return structuredClone(initialMonitors);
    },
    async add(monitor) {
      return await new Promise((resolve) => {
        socket.emit("add", monitor, (response) => {
          resolve(response ?? { ok: false, msg: "empty add response" });
        });
      });
    },
    async editMonitor(monitor) {
      return await new Promise((resolve) => {
        socket.emit("editMonitor", monitor, (response) => {
          resolve(response ?? { ok: false, msg: "empty editMonitor response" });
        });
      });
    },
    async deleteMonitor(monitorID) {
      return await new Promise((resolve) => {
        socket.emit("deleteMonitor", monitorID, (response) => {
          resolve(response ?? { ok: false, msg: "empty deleteMonitor response" });
        });
      });
    },
  };
}
