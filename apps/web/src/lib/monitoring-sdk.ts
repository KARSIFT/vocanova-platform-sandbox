// The only parts of the browser SDK the app uses. Loading this module instead
// of the whole package namespace lets the bundler leave the rest out of the
// deferred monitoring chunk.
export {
  captureException,
  captureRouterTransitionStart,
  init,
} from "@sentry/nextjs";
