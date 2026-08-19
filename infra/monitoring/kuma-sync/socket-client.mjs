import { io } from "socket.io-client";

import { authenticateAndLoadMonitors } from "./auth-handshake.mjs";

export {
  authenticateAndLoadMonitors,
  createMonitorListWaiter,
} from "./auth-handshake.mjs";

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

  try {
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

    const initialMonitors = await authenticateAndLoadMonitors(socket, {
      username,
      password,
      timeoutMs,
    });

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
  } catch (error) {
    socket.disconnect();
    throw error;
  }
}
