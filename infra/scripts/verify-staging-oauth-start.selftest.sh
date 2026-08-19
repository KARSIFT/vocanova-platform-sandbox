#!/usr/bin/env bash
set -euo pipefail

# VOC-084-T02 — disposable harness for verify-staging-oauth-start.sh.
#
# Runs the checker against a local fake API on 127.0.0.1. No access to the
# real staging host is required.
#
# Usage: infra/scripts/verify-staging-oauth-start.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
verify_script="$repo_root/infra/scripts/verify-staging-oauth-start.sh"
fake_server="$repo_root/infra/scripts/.verify-staging-oauth-start-fake-server.py"
port=8766
api_base="http://127.0.0.1:$port"
web_redirect="https://staging.vocanova.site/home"
canonical_callback="https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback"

cat > "$fake_server" <<PYEOF
import http.server
import json
import os
import sys
from urllib.parse import quote

scenario = os.environ.get("FAKE_SCENARIO", "disabled")
healthz_ready = os.environ.get("FAKE_HEALTHZ_READY", "")
canonical = "${canonical_callback}"
web_redirect = "${web_redirect}"

def healthz_controlled_signup_ready():
    if healthz_ready == "true":
        return True
    if healthz_ready == "false":
        return False
    return scenario == "enabled"

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        if self.path != "/healthz":
            self.send_response(404)
            self.end_headers()
            return

        body = json.dumps(
            {
                "status": "ok",
                "database": "ok",
                "controlled_signup_ready": healthz_controlled_signup_ready(),
                "kill_switches": {
                    "oauth_enabled": scenario == "enabled",
                    "new_signups_enabled": False,
                },
            }
        ).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_POST(self):
        if self.path != "/api/v1/auth/oauth/google/start":
            self.send_response(404)
            self.end_headers()
            return

        if scenario == "enabled":
            auth_url = (
                "https://accounts.google.com/o/oauth2/v2/auth?"
                f"client_id=test-client&redirect_uri={quote(canonical, safe='')}"
                "&response_type=code&scope=openid+email+profile&state=test-state"
            )
            body = json.dumps({"url": auth_url}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        if scenario == "wrong_callback":
            auth_url = (
                "https://accounts.google.com/o/oauth2/v2/auth?"
                "client_id=test-client&redirect_uri=https%3A%2F%2Fevil.example%2Fcb"
                "&response_type=code&scope=openid+email+profile&state=test-state"
            )
            body = json.dumps({"url": auth_url}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        if scenario == "lookalike_host":
            auth_url = (
                "https://accounts.google.com.attacker.example/o/oauth2/v2/auth?"
                f"client_id=test-client&redirect_uri={quote(canonical, safe='')}"
                "&response_type=code&scope=openid+email+profile&state=test-state"
            )
            body = json.dumps({"url": auth_url}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        self.send_response(503)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"detail":"Google OAuth sign-in is disabled"}')

http.server.HTTPServer(("127.0.0.1", int(sys.argv[1])), Handler).serve_forever()
PYEOF

start_server() {
  FAKE_SCENARIO="$1" FAKE_HEALTHZ_READY="${2:-}" python3 "$fake_server" "$port" &
  server_pid=$!
  for _ in $(seq 1 50); do
    status="$(curl -sS --max-time 1 -o /dev/null -w "%{http_code}" \
      -X POST -H "Content-Type: application/json" \
      -d "{\"redirectUri\":\"$web_redirect\"}" \
      "$api_base/api/v1/auth/oauth/google/start" 2>/dev/null || true)"
    healthz_status="$(curl -sS --max-time 1 -o /dev/null -w "%{http_code}" \
      "$api_base/healthz" 2>/dev/null || true)"
    if { [ "$status" = "200" ] || [ "$status" = "503" ]; } && [ "$healthz_status" = "200" ]; then
      return 0
    fi
    sleep 0.1
  done
  echo "fake server did not come up" >&2
  exit 1
}

stop_server() {
  kill "$server_pid" 2>/dev/null || true
  wait "$server_pid" 2>/dev/null || true
}

cleanup() {
  stop_server
  rm -f "$fake_server"
}
trap cleanup EXIT

failures=0
check() {
  local desc="$1" expect_pass="$2"; shift 2
  if "$@" >/tmp/verify-staging-oauth-selftest-output.txt 2>&1; then
    result=pass
  else
    result=fail
  fi
  if [ "$result" = "$expect_pass" ]; then
    echo "SELFTEST OK: $desc"
  else
    echo "SELFTEST FAILED: $desc (got $result, expected $expect_pass)" >&2
    cat /tmp/verify-staging-oauth-selftest-output.txt >&2
    failures=$((failures + 1))
  fi
  rm -f /tmp/verify-staging-oauth-selftest-output.txt
}

echo "== case 1: disabled OAuth returns 503 =="
start_server disabled
check "disabled OAuth start is refused with 503" pass \
  env EXPECT_OAUTH_ENABLED=false STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 2: enabled OAuth returns Google URL with canonical callback =="
start_server enabled
check "enabled OAuth start validates accounts.google.com and redirect_uri" pass \
  env EXPECT_OAUTH_ENABLED=true STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 3: wrong callback must fail when enabled =="
start_server wrong_callback
check "wrong redirect_uri fails the enabled check" fail \
  env EXPECT_OAUTH_ENABLED=true STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 4: fake 200 while disabled must fail =="
start_server enabled
check "enabled expectation against disabled server fails" fail \
  env EXPECT_OAUTH_ENABLED=false STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 5: lookalike Google hostname must fail =="
start_server lookalike_host
check "Google lookalike hostname fails the enabled check" fail \
  env EXPECT_OAUTH_ENABLED=true STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 6: enabled OAuth with controlled_signup_ready=false must fail =="
start_server enabled false
check "controlled_signup_ready=false fails the enabled readiness check" fail \
  env EXPECT_OAUTH_ENABLED=true STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

echo "== case 7: enabled OAuth with controlled_signup_ready=true must pass =="
start_server enabled true
check "controlled_signup_ready=true passes the enabled readiness check" pass \
  env EXPECT_OAUTH_ENABLED=true STAGING_API_HOST=127.0.0.1 \
  bash "$verify_script" "$api_base" "$web_redirect"
stop_server

if [ "$failures" -gt 0 ]; then
  echo "$failures selftest case(s) failed" >&2
  exit 1
fi
echo "ALL SELFTEST CASES PASSED"
