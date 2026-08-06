// `apps/web` source is written for Next.js's bundler resolution, so its imports
// omit file extensions (`./lib/env`, `next/server`). Node's ESM resolver requires
// them, so importing `src/middleware.ts` directly from `node --test` fails without
// this hook. Registering it keeps the middleware tests running against the real
// source file instead of a copy that would drift from it.
import { registerHooks } from "node:module";

const CANDIDATE_EXTENSIONS = [".ts", ".tsx", ".js", ".mjs", ".cjs"];

registerHooks({
  resolve(specifier, context, nextResolve) {
    try {
      return nextResolve(specifier, context);
    } catch (originalError) {
      for (const extension of CANDIDATE_EXTENSIONS) {
        try {
          return {
            ...nextResolve(`${specifier}${extension}`, context),
            shortCircuit: true,
          };
        } catch {
          continue;
        }
      }
      throw originalError;
    }
  },
});
