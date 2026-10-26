#!/usr/bin/env python3
"""Real nginx regression using only disposable containers and synthetic cookies."""
import argparse
import http.client
import ipaddress
import json
from pathlib import Path
import shutil
import ssl
import subprocess
import sys
import tempfile
import time
import uuid

IMAGE = "nginx:1.27-alpine"
LABEL = "vocanova.auth-proxy-regression"
HOSTS = {
    "staging.vocanova.site": ("staging", "web"),
    "api-staging.vocanova.site": ("staging", "api"),
    "production.vocanova.site": ("production", "web"),
    "api-production.vocanova.site": ("production", "api"),
    "monitor.vocanova.site": (None, "monitor"),
}
ALIASES = ("vocanova-web", "vocanova-api", "vocanova-production-web",
           "vocanova-production-api", "vocanova-uptime-kuma")
FORGED = {"198.51.100.21", "198.51.100.22", "198.51.100.23", "198.51.100.24"}


def run(*args, check=True, timeout=20):
    result = subprocess.run(args, capture_output=True, text=True, timeout=timeout)
    if check and result.returncode:
        detail = (result.stderr or result.stdout).strip()[-1600:]
        raise RuntimeError(f"{args[0]} failed ({result.returncode}): {detail}")
    return result


def cleanup(kind, name, owner):
    labels = ".Config.Labels" if kind == "container" else ".Labels"
    template = "{{index " + labels + " " + json.dumps(LABEL) + "}}"
    found = run("docker", kind, "inspect", "--format", template, name, check=False)
    if found.returncode:
        return
    if found.stdout.strip() != owner:
        raise RuntimeError("Refusing cleanup of an unowned resource")
    if kind == "container":
        run("docker", "rm", "-f", name)
    else:
        run("docker", "network", "rm", name)


def fixture(root, directory):
    fragments = directory / "fragments"
    fragments.mkdir()
    sources = {
        "00-ip.conf": "infra/nginx-shared/conf.d/00-cloudflare-real-ip.conf",
        "01-tls.conf": "infra/nginx-shared/conf.d/01-tls.conf",
        "02-dns.conf": "infra/nginx-shared/conf.d/02-docker-dns.conf",
        "10-staging-web.conf": "infra/nginx/conf.d/10-staging-web.conf",
        "20-staging-api.conf": "infra/nginx/conf.d/20-api-staging.conf",
        "30-production-web.conf": "infra/nginx-production/conf.d/10-production-web.conf",
        "40-production-api.conf": "infra/nginx-production/conf.d/20-api-production.conf",
        "50-monitor.conf": "infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf",
    }
    for destination, source in sources.items():
        text = (root / source).read_text(encoding="utf-8")
        text = text.replace("__PRODUCTION_WEB_HOST__", "production.vocanova.site")
        text = text.replace("__PRODUCTION_API_HOST__", "api-production.vocanova.site")
        (fragments / destination).write_text(text, encoding="utf-8")
    certs = directory / "certificates"
    certs.mkdir()
    run("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes", "-days", "1",
        "-subj", "/CN=auth-proxy-fixture.invalid", "-keyout", str(certs / "key.pem"),
        "-out", str(certs / "cert.pem"))
    (certs / "key.pem").chmod(0o600)
    (directory / "nginx.conf").write_text(r"""
worker_processes 1;
pid /tmp/nginx.pid;
error_log /dev/stderr warn;
events { worker_connections 128; }
http {
    log_format regression escape=json
        '{"port":"$server_port","host":"$host","uri":"$request_uri","peer":"$remote_addr"}';
    access_log /tmp/proxy-regression.jsonl regression;
    client_body_temp_path /tmp/client;
    proxy_temp_path /tmp/proxy;
    fastcgi_temp_path /tmp/fastcgi;
    uwsgi_temp_path /tmp/uwsgi;
    scgi_temp_path /tmp/scgi;
    include /fixture/fragments/*.conf;
    server {
        listen 3000;
        listen 8080;
        listen 3001;
        location / {
            add_header X-Test-Cookie $http_cookie always;
            add_header X-Test-Real-IP $http_x_real_ip always;
            add_header X-Test-XFF $http_x_forwarded_for always;
            return 200 "synthetic proxy echo\n";
        }
    }
}
""".lstrip(), encoding="utf-8")


def request(port, host, uri, cookie):
    # This unverified certificate belongs only to the loopback fixture.
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE
    connection = http.client.HTTPSConnection("127.0.0.1", port, context=context, timeout=8)
    headers = {
        "Host": host, "X-Real-IP": "198.51.100.21",
        "X-Forwarded-For": "198.51.100.22, 198.51.100.23",
        "CF-Connecting-IP": "198.51.100.24",
    }
    if cookie:
        headers["Cookie"] = cookie
    try:
        connection.request("GET", uri, headers=headers)
        response = connection.getresponse()
        response.read()
        if response.status != 200:
            raise AssertionError(f"{host} returned {response.status}, expected echo backend")
        return tuple(response.getheader(name) for name in
                     ("X-Test-Cookie", "X-Test-Real-IP", "X-Test-XFF"))
    finally:
        connection.close()


def verify(container, port):
    values = {
        "vocanova_session": "legacy-session-fixture", "vocanova_csrf": "legacy-csrf-fixture",
        "unrelated": "not-application-state", "vocanova_production_session": "production-fixture",
        "vocanova_staging_csrf": "staging-csrf-fixture", "vocanova_theme": "dark",
        "vocanova_oauth_state": "oauth-fixture", "vocanova_staging_session": "staging-fixture",
        "vocanova_production_csrf": "production-csrf-fixture",
    }
    results = []
    for host, (tier, kind) in HOSTS.items():
        allowed = set() if tier is None else {
            f"vocanova_{tier}_session", f"vocanova_{tier}_csrf",
            "vocanova_theme" if kind == "web" else "vocanova_oauth_state",
        }
        cases = {
            "mixed": values,
            "foreign-only": {k: v for k, v in values.items() if k not in allowed},
            "absent": {},
        }
        for case, supplied in cases.items():
            uri = f"/auth-proxy-regression/{case}"
            header = "; ".join(f"{k}={v}" for k, v in supplied.items())
            actual_cookie, real_ip, xff = request(port, host, uri, header)
            pairs = {}
            for entry in (actual_cookie or "").split(";"):
                if not entry.strip():
                    continue
                name, separator, value = entry.strip().partition("=")
                if not separator or name in pairs or name not in allowed:
                    raise AssertionError(f"{host}/{case} forwarded an unapproved cookie")
                pairs[name] = value
            # Explicit proxy headers may include empty own-name fields.
            nonempty = {k: v for k, v in pairs.items() if v}
            expected = {k: v for k, v in supplied.items() if k in allowed}
            if nonempty != expected:
                raise AssertionError(f"{host}/{case} changed or leaked cookie values")
            if kind == "monitor" and actual_cookie is not None:
                raise AssertionError("Monitor must omit upstream Cookie entirely")
            results.append((host, uri, real_ip, xff))
    log = run("docker", "exec", container, "cat", "/tmp/proxy-regression.jsonl").stdout
    peers = {}
    for line in log.splitlines():
        record = json.loads(line)
        if record["port"] == "443":
            peers[(record["host"], record["uri"])] = record["peer"]
    for host, uri, real_ip, xff in results:
        peer = peers.get((host, uri))
        if peer is None:
            raise AssertionError(f"Missing actual nginx peer for {host}{uri}")
        ipaddress.ip_address(peer)
        if peer in FORGED or real_ip != peer or xff != peer:
            raise AssertionError(f"{host}{uri} failed actual-peer forwarding")
    return len(results)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--require-docker", action="store_true")
    args = parser.parse_args()
    available = False
    if shutil.which("docker"):
        try:
            available = run("docker", "info", "--format", "{{.ServerVersion}}",
                            check=False, timeout=15).returncode == 0
        except (OSError, subprocess.TimeoutExpired):
            pass
    if not available:
        message = "Docker unavailable; auth proxy container regression was not run"
        if args.require_docker:
            raise RuntimeError(message)
        print("SKIP: " + message)
        return
    if not shutil.which("openssl"):
        raise RuntimeError("openssl required for disposable TLS fixture")
    root = Path(__file__).resolve().parents[2]
    owner = uuid.uuid4().hex
    container, network = f"vocanova-auth-proxy-{owner}", f"vocanova-auth-proxy-net-{owner}"
    with tempfile.TemporaryDirectory(prefix="vocanova-auth-proxy-") as temporary:
        directory = Path(temporary)
        fixture(root, directory)
        try:
            if run("docker", "image", "inspect", IMAGE, check=False).returncode:
                run("docker", "pull", IMAGE, timeout=45)
            run("docker", "network", "create", "--label", f"{LABEL}={owner}", network)
            command = [
                "docker", "create", "--name", container, "--label", f"{LABEL}={owner}",
                "--network", network, "--read-only", "--tmpfs", "/tmp:rw,noexec,nosuid,size=16m",
                "--publish", "127.0.0.1::443", "--mount",
                f"type=bind,source={directory},target=/fixture,readonly",
            ]
            for tier in ("staging", "production"):
                command += ["--mount", f"type=bind,source={directory / 'certificates'},target=/etc/nginx/certs/{tier},readonly"]
            for alias in ALIASES:
                command += ["--network-alias", alias]
            command += ["--entrypoint", "nginx", IMAGE, "-c", "/fixture/nginx.conf", "-g", "daemon off;"]
            run(*command)
            run("docker", "start", container)
            run("docker", "exec", container, "nginx", "-t", "-c", "/fixture/nginx.conf")
            binding = run("docker", "port", container, "443/tcp").stdout.strip()
            if not binding.startswith("127.0.0.1:") or "\n" in binding:
                raise RuntimeError("Fixture listener is not exclusively loopback")
            port = int(binding.rsplit(":", 1)[1])
            for attempt in range(20):
                try:
                    request(port, "staging.vocanova.site", "/auth-proxy-regression/ready", None)
                    break
                except (OSError, http.client.HTTPException):
                    if attempt == 19:
                        raise
                    time.sleep(0.1)
            count = verify(container, port)
            print(f"PASS: actual nginx -t and {count} cookie/IP cases across five hosts")
        finally:
            # Exact ownership labels: never prune, remove images, or touch live Compose.
            try:
                cleanup("container", container, owner)
            finally:
                cleanup("network", network, owner)


if __name__ == "__main__":
    try:
        main()
    except (AssertionError, RuntimeError, OSError, ValueError, subprocess.TimeoutExpired) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        sys.exit(1)

