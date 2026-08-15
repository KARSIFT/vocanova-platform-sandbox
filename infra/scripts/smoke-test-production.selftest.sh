#!/usr/bin/env bash
set -euo pipefail

# VOC-038-T02 — disposable rehearsal harness for
# smoke-test-production.sh, mirroring the existing
# rehearse-production-secrets-boundary.selftest.sh convention: a
# script that only ever runs against a correctly configured target
# proves nothing about the checker itself, so this harness runs it
# against a local fake server in both a healthy and several broken
# configurations and confirms it passes/fails as expected in each.
# VOC-085-T01 — empty-content rejection and detail-endpoint fixtures.
# VOC-085-T02 — route-sweep coverage, auth-cookie handling, fail-closed.
#
# Requires: python3, curl, bash. Runs entirely on 127.0.0.1; no
# network access to any real host.
#
# Usage: infra/scripts/smoke-test-production.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
smoke_script="$repo_root/infra/scripts/smoke-test-production.sh"
fake_server="$repo_root/infra/scripts/.smoke-test-fake-server.py"
port=8765
base_url="http://127.0.0.1:$port"

cat > "$fake_server" <<'PYEOF'
import http.server
import json
import os
import sys

scenario = os.environ.get("FAKE_SCENARIO", "healthy")
SMOKE_COOKIE = "vocanova_session=smoke-test-token"

SITUATION_LIST = {
    "items": [
        {
            "id": "00000000-0000-0000-0000-000000000001",
            "slug": "airport",
            "title": "Airport",
            "shortDescription": "Airport words.",
            "category": "travel",
            "displayOrder": 1,
        }
    ],
    "nextCursor": "",
}

SITUATION_DETAIL = {
    "situation": {
        "id": "00000000-0000-0000-0000-000000000001",
        "slug": "airport",
        "title": "Airport",
        "shortDescription": "Airport words.",
        "category": "travel",
        "displayOrder": 1,
    },
    "meanings": [
        {
            "meaningId": "00000000-0000-0000-0000-000000000003",
            "wordId": "00000000-0000-0000-0000-000000000002",
            "wordSlug": "boarding-pass",
            "wordText": "boarding pass",
            "partOfSpeech": "noun",
            "shortDefinition": "A document that lets you get on your flight.",
            "saved": False,
        }
    ],
}

WORD_DETAIL = {
    "word": {
        "id": "00000000-0000-0000-0000-000000000002",
        "text": "boarding pass",
        "slug": "boarding-pass",
        "wordType": "phrase",
        "meanings": [
            {
                "id": "00000000-0000-0000-0000-000000000003",
                "partOfSpeech": "noun",
                "shortDefinition": "A document that lets you get on your flight.",
                "saved": False,
                "examples": [],
                "usageNotes": [],
            }
        ],
    }
}

PUBLIC_WEB_PATHS = {"/", "/signin", "/auth/magic"}
AUTHENTICATED_WEB_PATHS = {
    "/onboarding",
    "/home",
    "/discover",
    "/reviews",
    "/progress",
    "/settings",
    "/settings/account",
    "/discover/airport",
    "/discover/airport/boarding-pass",
}

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def _json(self, status, body):
        payload = json.dumps(body).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def _authorized(self):
        return self.headers.get("Cookie", "") == SMOKE_COOKIE

    def _serve_web(self, path):
        if path in PUBLIC_WEB_PATHS:
            self.send_response(200)
            self.end_headers()
            return
        if path in AUTHENTICATED_WEB_PATHS:
            if scenario == "route_auth_failure":
                self.send_response(500)
                self.end_headers()
                return
            if self._authorized():
                if path == "/onboarding":
                    self.send_response(307)
                    self.send_header("Location", "/home")
                    self.end_headers()
                    return
                self.send_response(200)
                self.end_headers()
                return
            self.send_response(302)
            self.send_header("Location", "/signin?returnTo=" + path)
            self.end_headers()
            return
        self.send_response(404)
        self.end_headers()

    def do_GET(self):
        if self.path == "/healthz":
            switches = {
                "magic_link_enabled": False,
                "oauth_enabled": False,
                "new_signups_enabled": False,
                "ai_enabled": True,
            }
            status = "ok"
            database = "ok"
            if scenario == "db_unhealthy":
                status = "unhealthy"
                database = "unhealthy"
            if scenario == "switch_mismatch":
                switches["ai_enabled"] = False
            if scenario == "switches_on":
                switches["magic_link_enabled"] = True
                switches["oauth_enabled"] = True
            self._json(200, {"status": status, "database": database, "kill_switches": switches})
        elif self.path in PUBLIC_WEB_PATHS or self.path in AUTHENTICATED_WEB_PATHS:
            self._serve_web(self.path)
        elif self.path == "/api/v1/me":
            if self._authorized():
                self._json(200, {"ok": True})
            else:
                self._json(401, {"error": "unauthorized"})
        elif self.path == "/api/v1/journey-situations":
            if not self._authorized():
                self._json(401, {"error": "unauthorized"})
            elif scenario == "empty_situations":
                self._json(200, {"items": [], "nextCursor": ""})
            else:
                self._json(200, SITUATION_LIST)
        elif self.path == "/api/v1/journey-situations/airport":
            if not self._authorized():
                self._json(401, {"error": "unauthorized"})
            elif scenario == "missing_word_detail":
                self._json(200, {"situation": SITUATION_DETAIL["situation"], "meanings": []})
            else:
                self._json(200, SITUATION_DETAIL)
        elif self.path == "/api/v1/canonical-words/boarding-pass":
            if not self._authorized():
                self._json(401, {"error": "unauthorized"})
            elif scenario == "word_detail_missing":
                self._json(404, {"error": "not found"})
            else:
                self._json(200, WORD_DETAIL)
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == "/api/v1/auth/magic-links":
            self.send_response(503 if scenario != "switches_on" else 202)
            self.end_headers()
        elif self.path == "/api/v1/auth/oauth/google/start":
            self.send_response(503 if scenario != "switches_on" else 200)
            self.end_headers()
        else:
            self.send_response(404)
            self.end_headers()

http.server.HTTPServer(("127.0.0.1", int(sys.argv[1])), Handler).serve_forever()
PYEOF

start_server() {
  FAKE_SCENARIO="$1" python3 "$fake_server" "$port" &
  server_pid=$!
  for _ in $(seq 1 30); do
    if curl -fsS --max-time 1 "$base_url/healthz" >/dev/null 2>&1; then
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
  if "$@" >/tmp/smoke-selftest-output.txt 2>&1; then
    result=pass
  else
    result=fail
  fi
  if [ "$result" = "$expect_pass" ]; then
    echo "SELFTEST OK: $desc"
  else
    echo "SELFTEST FAILED: $desc (got $result, expected $expect_pass)" >&2
    cat /tmp/smoke-selftest-output.txt >&2
    failures=$((failures + 1))
  fi
  rm -f /tmp/smoke-selftest-output.txt
}

echo "== case 1: healthy target, defaults match =="
start_server healthy
check "healthy target passes with default expectations and route sweep" pass \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 2: database unhealthy must fail =="
start_server db_unhealthy
check "unhealthy database fails the suite" fail \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 3: kill-switch mismatch must fail =="
start_server switch_mismatch
check "kill-switch mismatch fails the suite" fail \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 4: switches reported on, expectations set to match =="
start_server switches_on
check "matching non-default expectations pass" pass \
  env EXPECT_MAGIC_LINK_ENABLED=true EXPECT_OAUTH_ENABLED=true \
      SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 5: core-loop check runs and passes when a session cookie is supplied =="
start_server healthy
check "core-loop check passes with a valid smoke-test cookie" pass \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 6: empty journey-situations list must fail with a session cookie =="
start_server empty_situations
check "empty journey-situations body fails the suite" fail \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 7: situation detail without meanings must fail =="
start_server missing_word_detail
check "missing situation meanings fails the suite" fail \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 8: missing canonical word detail must fail =="
start_server word_detail_missing
check "missing canonical word detail fails the suite" fail \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 9: route sweep fails closed without a session cookie =="
start_server healthy
check "missing session cookie fails route sweep (not silent skip)" fail \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 10: invalid session cookie fails authenticated route sweep =="
start_server healthy
check "malformed session cookie fails authenticated route coverage" fail \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=not-the-smoke-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

echo "== case 11: protected route server error fails route sweep =="
start_server route_auth_failure
check "protected route HTTP 500 fails route sweep" fail \
  env SMOKE_TEST_SESSION_COOKIE="vocanova_session=smoke-test-token" \
  bash "$smoke_script" "$base_url" "$base_url"
stop_server

if [ "$failures" -gt 0 ]; then
  echo "$failures selftest case(s) failed" >&2
  exit 1
fi
echo "ALL SELFTEST CASES PASSED"
