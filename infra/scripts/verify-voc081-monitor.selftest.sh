#!/usr/bin/env bash
set -euo pipefail

# VOC-081-T04 — disposable harness for verify-voc081-monitor.sh.
#
# Runs the production script against a local fake Kuma login + socket.io server
# on 127.0.0.1. No access to the real monitor hostname is required.
#
# Usage: infra/scripts/verify-voc081-monitor.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
verify_script="$repo_root/infra/scripts/verify-voc081-monitor.sh"
test_root="$(mktemp -d)"
fake_server="$test_root/fake-server.py"
port=8871
base_url="http://127.0.0.1:$port"

cat > "$fake_server" <<'PYEOF'
import http.server
import os
import sys

scenario = os.environ.get("FAKE_SCENARIO", "healthy")

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        if self.path.startswith("/socket.io/"):
            if scenario == "healthy":
                if "transport=polling" in self.path:
                    body = b'0{"sid":"test","upgrades":["websocket"],"pingInterval":25000,"pingTimeout":20000}'
                    self.send_response(200)
                    self.send_header("Content-Type", "text/plain")
                    self.send_header("Content-Length", str(len(body)))
                    self.end_headers()
                    self.wfile.write(body)
                    return
                self.send_response(101)
                self.send_header("Upgrade", "websocket")
                self.send_header("Connection", "Upgrade")
                self.end_headers()
                return
            self.send_response(502)
            self.end_headers()
            self.wfile.write(b"upstream down")
            return

        if self.path.startswith("/api/entry-page"):
            if scenario == "anonymous_dashboard":
                body = b'{"type":"entryPage","token":"leaked","authenticated":true,"username":"admin"}'
            else:
                body = b'{"type":"entryPage","entryPage":null}'
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        if scenario == "healthy":
            # Simulate Kuma: / → 302 /dashboard SPA shell
            if self.path in ("/", ""):
                self.send_response(302)
                self.send_header("Location", "/dashboard")
                self.end_headers()
                self.wfile.write(b"Found. Redirecting to /dashboard")
                return
            body = (
                b"<html><head><title>Uptime Kuma</title>"
                b"<meta name='description' content='Uptime Kuma monitoring tool' />"
                b"</head><body><div id='app'></div></body></html>"
            )
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        if scenario == "anonymous_dashboard":
            body = b"<html><title>Admin</title><body>Dashboard Add New Monitor</body></html>"
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return

        self.send_response(502)
        self.end_headers()
        self.wfile.write(b"error")

if __name__ == "__main__":
    port = int(sys.argv[1])
    http.server.HTTPServer(("127.0.0.1", port), Handler).serve_forever()
PYEOF

server_pid=""
stop_server() {
  if [ -n "$server_pid" ]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
    server_pid=""
  fi
}
final_cleanup() {
  stop_server
  rm -rf "$test_root"
}
trap final_cleanup EXIT

start_server() {
  local scenario="$1"
  stop_server
  FAKE_SCENARIO="$scenario" python3 "$fake_server" "$port" &
  server_pid=$!
  sleep 0.3
}

run_expect() {
  local label="$1"
  local expect_exit="$2"
  local scenario="$3"
  start_server "$scenario"
  set +e
  MONITOR_BASE_URL="$base_url" "$verify_script" --skip-apps
  local exit_code=$?
  set -e
  if [ "$exit_code" -ne "$expect_exit" ]; then
    echo "SELFTEST FAIL: $label expected exit $expect_exit, got $exit_code" >&2
    exit 1
  fi
  echo "SELFTEST PASS: $label"
}

run_expect "healthy monitor login + websocket" 0 "healthy"
run_expect "origin 502 is rejected" 1 "upstream_502"
run_expect "anonymous dashboard is rejected" 1 "anonymous_dashboard"

echo "All verify-voc081-monitor.selftest checks passed."
