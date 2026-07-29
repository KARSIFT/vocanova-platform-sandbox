#!/usr/bin/env sh
# infra/nginx/generate-dev-cert.sh
#
# VOC-032-T05: generates a throwaway self-signed certificate/key
# pair at infra/secrets/nginx/{cert,key}.pem, substituting for the
# real Cloudflare origin certificate (VOC-032-DEP-01, not yet
# provisioned) so `docker compose up` and `nginx -t` can be
# exercised locally (VOC-032-TEST-11, VOC-032-TEST-12).
#
# This is a LOCAL-VALIDATION-ONLY convenience. It never runs in CI
# and is never used on the real staging host - the founder replaces
# infra/secrets/nginx/{cert,key}.pem with the real Cloudflare origin
# certificate there (infra/secrets/.gitignore already keeps this
# directory out of git, so nothing this script writes is ever
# committed).
#
# Usage: sh infra/nginx/generate-dev-cert.sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
out_dir="$script_dir/../secrets/nginx"
mkdir -p "$out_dir"

openssl req -x509 -nodes -newkey rsa:2048 \
    -keyout "$out_dir/key.pem" \
    -out "$out_dir/cert.pem" \
    -days 30 \
    -subj "/CN=vocanova.site" \
    -addext "subjectAltName=DNS:staging.vocanova.site,DNS:api-staging.vocanova.site"

echo "Wrote self-signed dev cert/key to $out_dir (never committed - see infra/secrets/.gitignore)"
