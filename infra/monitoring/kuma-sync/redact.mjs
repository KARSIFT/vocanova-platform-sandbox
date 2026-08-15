const SECRET_PATTERNS = [
  /(password\s*[:=]\s*)([^\s"']+)/gi,
  /(KUMA_PASSWORD\s*[:=]\s*)([^\s"']+)/gi,
  /(token\s*[:=]\s*)([^\s"']+)/gi,
  /(authorization\s*[:=]\s*)([^\s"']+)/gi,
  /(Bearer\s+)([A-Za-z0-9._-]+)/gi,
];

export function redactSecrets(text) {
  let output = String(text);
  for (const pattern of SECRET_PATTERNS) {
    output = output.replace(pattern, "$1[REDACTED]");
  }
  return output;
}

export function createRedactingLogger(logger = console) {
  return {
    info(message) {
      logger.info(redactSecrets(message));
    },
    error(message) {
      logger.error(redactSecrets(message));
    },
  };
}
