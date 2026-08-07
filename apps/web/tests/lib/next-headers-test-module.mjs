const COOKIE_HEADER_GLOBAL_KEY = "__voc_test_cookie_header";

export async function headers() {
  const cookieHeader = globalThis[COOKIE_HEADER_GLOBAL_KEY];

  return {
    get(name) {
      if (name.toLowerCase() === "cookie") {
        return typeof cookieHeader === "string" ? cookieHeader : null;
      }
      return null;
    },
  };
}
