#!/bin/bash
# Smoke test for the update-apex-ips image, run by .github/workflows/publish.yml on the image it
# built and again on the image it pushed, and runnable by hand:
#   update-apex-ips/smoke-test.sh <image> [expected-version]
# Each assertion is its own line so any failure stops the script. Nothing here touches the network
# or AWS.
set -euo pipefail

image=${1:?usage: smoke-test.sh <image> [expected-version]}
want=${2:-}

# It starts and documents its flags.
usage=$(docker run --rm "$image" -h 2>&1)
grep -q -- '-name' <<<"$usage"
grep -q -- '-apply' <<<"$usage"

# It runs as the distroless nonroot user, not root.
[ "$(docker inspect --format '{{.Config.User}}' "$image")" = 65532 ]

# The image holds the release it is tagged for: `go install ...@<tag>` embeds the module version.
version=$(docker run --rm "$image" -version)
[ -n "$version" ]
[ -z "$want" ] || [ "$version" = "$want" ]

# It refuses an apex name before doing any lookup, with a clear message and exit code 1.
rc=0
out=$(docker run --rm "$image" -name example.com 2>&1) || rc=$?
[ "$rc" = 1 ]
grep -q 'already the apex' <<<"$out"

echo "update-apex-ips smoke test OK (version $version)"
