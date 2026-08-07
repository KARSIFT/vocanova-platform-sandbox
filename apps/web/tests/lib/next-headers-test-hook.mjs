import { registerHooks } from "node:module";

const MOCK_HEADERS_MODULE_URL = new URL(
  "./next-headers-test-module.mjs",
  import.meta.url,
).href;

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (specifier === "next/headers") {
      return {
        url: MOCK_HEADERS_MODULE_URL,
        shortCircuit: true,
      };
    }
    return nextResolve(specifier, context);
  },
});
