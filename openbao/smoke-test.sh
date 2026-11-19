#!/bin/bash
# Smoke test for the openbao image, run by .github/workflows/publish.yml on the image it built and
# again on the image it pushed, and runnable by hand:
#   openbao/smoke-test.sh <image> [tag]
# Each assertion is its own line, so any failure stops the script whatever its position. The
# `plugins` list repeats the Dockerfile's checksums on purpose: it is an independent check of
# what was built, as does `helper`. Nothing here touches the network or runs a plugin; the
# helper runs only to print its version.
set -euo pipefail

image=${1:?usage: smoke-test.sh <image> [tag]}
tag=${2:-}

# file name inside /openbao/plugins : sha256 of the binary
plugins='kms-awskms-v0.1.1:0dffe0a159a5bdbeb147a892595fe517de858f48a99a64f33fabc775af0eec76
secret-aws-v0.3.2:c828ea1fd34a9d2c2c5546346a659216b871a5774a1beaa306061161f68a0a30'
# the Roles Anywhere credential helper: path : version : sha256
helper='/usr/local/bin/aws_signing_helper:1.8.6:3553fdc515e5682541a9ebdd37031ec7f461169aa13ccbb6fbd5b7259644113d'
IFS=: read -r helper_path helper_version helper_sha <<<"$helper"

list=$(tr '\n' ' ' <<<"$plugins")
# shellcheck disable=SC2086  # the list is split into words on purpose
want=$(set -- $list; echo $#)
[ "$want" -gt 0 ]

# The OpenBao version is the first part of the tag (2.7.2-kms-aws-...).
version=$(docker run --rm --entrypoint bao "$image" version)
grep -q '^OpenBao v' <<<"$version"
[ -z "$tag" ] || grep -q "^OpenBao v${tag%%-*} " <<<"$version"

docker run --rm --entrypoint sh "$image" -c '
  set -e
  n=0
  for p in '"$list"'; do
    f=/openbao/plugins/${p%%:*}
    ls -l "$f"
    echo "${p##*:}  $f" | sha256sum -c -
    [ "$(stat -c %U "$f")" = openbao ]
    [ -f "$f" ]
    [ ! -L "$f" ]
    [ -x "$f" ]
    n=$((n + 1))
  done
  [ "$n" -eq '"$want"' ]'

docker run --rm --entrypoint sh "$image" -c '
  set -e
  f='"$helper_path"'
  ls -l "$f"
  echo "'"$helper_sha"'  $f" | sha256sum -c -
  [ "$(stat -c %U "$f")" = root ]
  [ -f "$f" ]
  [ ! -L "$f" ]
  [ -x "$f" ]
  [ ! -w "$f" ]'
helper_out=$(docker run --rm --entrypoint "$helper_path" "$image" version)
echo "aws_signing_helper version: $helper_out"
grep -qx "$helper_version" <<<"$helper_out"

echo "openbao smoke test OK ($(head -1 <<<"$version" | cut -d' ' -f1-2))"
