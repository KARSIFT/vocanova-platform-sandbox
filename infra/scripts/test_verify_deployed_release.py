"""Exercise the real HTTP checker through a server that rejects generic clients."""
import importlib.util
import json
import os
from pathlib import Path
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("release_check", Path(__file__).with_name("verify-deployed-release.py"))
release_check = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release_check)


class ReleaseHTTPTest(unittest.TestCase):
    def test_named_client_verifies_both_services_through_bot_filter(self):
        requests = []
        identity = {"version": "0.2.0", "commit": "a" * 40, "environment": "staging", "builtAt": "2026-09-12T00:00:00.000Z"}

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                agent = self.headers.get("User-Agent", "")
                requests.append((self.path, agent))
                self.send_response(403 if not agent or agent.startswith("Python") else 200)
                self.end_headers()
                self.wfile.write(json.dumps(identity).encode())

            def log_message(self, *_args):
                pass

        server = HTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        url = f"http://127.0.0.1:{server.server_port}"
        try:
            with patch.dict(os.environ, {
                "RELEASE_VERSION": "0.2.0", "RELEASE_COMMIT": "a" * 40,
                "RELEASE_ENVIRONMENT": "staging", "RELEASE_BUILT_AT": "2026-09-12T00:00:00Z",
                "RELEASE_API_URL": url, "RELEASE_WEB_URL": url,
            }), patch.object(release_check.time, "sleep"):
                release_check.main()
            self.assertEqual(len(requests), 2)
            self.assertTrue(all(path == "/version" for path, _ in requests))
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

    def test_mixed_or_missing_release_identity_is_rejected(self):
        expected = {"version": "0.2.0", "commit": "a" * 40, "environment": "staging", "builtAt": "2026-09-12T00:00:00Z"}
        self.assertTrue(release_check.matches_identity(expected, expected))
        for key in expected:
            self.assertFalse(release_check.matches_identity({**expected, key: "unknown"}, expected))
        self.assertFalse(release_check.matches_identity({}, expected))
