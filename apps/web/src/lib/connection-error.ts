import { brand, neutral } from "@vocanova/design-tokens";

import { normalizeReturnTo } from "./return-to";
import { isThemePreference } from "./theme-preference";

export function connectionRetryDestination(returnTo?: string): string {
  const destination = normalizeReturnTo(returnTo);
  return new URL(destination, "https://vocanova.invalid").pathname ===
    "/connection-error"
    ? "/home"
    : destination;
}

function escapeAttribute(value: string): string {
  return value.replace(
    /[&<>"']/g,
    (character) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      })[character]!,
  );
}

// This deliberately needs neither the App Router nor JavaScript. An internal
// rewrite renders a page with a fresh 200 status; returning the document itself
// preserves the actual 503 and prevents any protected page from rendering.
export function connectionErrorDocument(
  returnTo: string,
  theme?: string,
): string {
  const preference = isThemePreference(theme) ? theme : "system";
  const destination = escapeAttribute(connectionRetryDestination(returnTo));
  const dark = `color-scheme:dark;--background:${neutral["900"]};--surface:${neutral["800"]};--text:${neutral["50"]};--muted:${neutral["300"]};--border:${neutral["700"]};--link:${brand.primary!["300"]}`;
  return `<!doctype html>
<html lang="en" data-theme-preference="${preference}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<title>Connection interrupted — Vocanova</title>
<style>
:root{color-scheme:light;--background:${neutral["100"]};--surface:${neutral["50"]};--text:${neutral["900"]};--muted:${neutral["700"]};--border:${neutral["200"]};--link:${brand.primary!["700"]}}
:root[data-theme-preference="dark"]{${dark}}
@media(prefers-color-scheme:dark){:root[data-theme-preference="system"]{${dark}}}
*{box-sizing:border-box}
body{margin:0;background:var(--background);color:var(--text);font-family:"Avenir Next",Avenir,"Segoe UI",system-ui,sans-serif;line-height:1.5}
main{max-width:36rem;margin:0 auto;padding:clamp(2rem,10vh,6rem) 1.25rem}
a{color:var(--link);display:inline-flex;align-items:center;min-height:44px;font-weight:600;text-underline-offset:4px;overflow-wrap:anywhere}
a:focus-visible{outline:3px solid var(--link);outline-offset:4px;border-radius:4px}
.brand{font-size:1.25rem;margin-bottom:1.5rem;text-decoration:none}
section{background:var(--surface);border:1px solid var(--border);border-radius:0.9rem;padding:clamp(1.25rem,5vw,2rem)}
h1{font-size:1.5rem;line-height:1.3;margin:0 0 1rem}p{color:var(--muted);margin:0 0 1.25rem}
</style>
</head>
<body><main><a class="brand" href="/" aria-label="VocaNova home">VocaNova</a>
<section aria-labelledby="connection-heading"><h1 id="connection-heading">We couldn't connect</h1>
<p>Vocanova is temporarily unavailable. Try again in a moment to continue.</p>
<a href="${destination}">Try again</a></section></main></body></html>`;
}
