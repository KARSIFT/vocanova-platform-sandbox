import base from "../../apps/web/playwright.cross-browser.config";

const directory = new URL(".", import.meta.url).pathname;
const webDirectory = new URL("../../apps/web/", import.meta.url).pathname;
const servers = Array.isArray(base.webServer)
  ? base.webServer
  : base.webServer
    ? [base.webServer]
    : [];

export default {
  ...base,
  testDir: directory,
  testMatch: "sentence-cookie-read-diagnostic.spec.ts",
  outputDir: webDirectory + "test-results/csrf-diagnostic",
  webServer: servers.map((server) => ({ ...server, cwd: webDirectory })),
  retries: 0,
  projects: base.projects!.filter((project) => project.name === "mobile-360"),
  use: { ...base.use, trace: "on", video: "off" },
};
