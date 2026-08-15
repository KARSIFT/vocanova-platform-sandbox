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
