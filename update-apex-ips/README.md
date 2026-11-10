# update-apex-ips

Keeps a zone apex A record in step with a subdomain's IP addresses in Route53. A zone apex can't be
a CNAME, so when `www.example.com` points at a load balancer by hostname, `example.com` needs the
same IPs as plain A records; this copies them over when they change.

- Source and documentation: <https://github.com/410Labs/update-apex-ips> (MIT)
- Image: `410labs/update-apex-ips:<tag>` on Docker Hub, public, `linux/amd64` and `linux/arm64`.
  The tag is the release of the tool it contains (`go install ...@<tag>`); `-version` prints it.
  Tags are never overwritten and there is no `latest`: pin `<tag>@<digest>` (the digest is in the
  summary of the publish run).

```sh
# plan only: compares www.example.com with the apex and prints the difference (exit 2 on drift)
docker run --rm -e AWS_REGION=us-east-1 \
  -v "$PWD/credentials:/aws/credentials:ro" -e AWS_SHARED_CREDENTIALS_FILE=/aws/credentials \
  410labs/update-apex-ips:v0.1.0 -name www.example.com

# update Route53 (UPSERT of the apex A record only)
... 410labs/update-apex-ips:v0.1.0 -name www.example.com -apply
```

Credentials come from the standard AWS chain. Exit codes: `0` in sync (or applied), `1` error,
`2` drift found in plan-only mode.

## Building and bumping

CI builds and publishes this image (`.github/workflows/publish.yml`, see the repository README).
`TAG` holds the release to publish and `PLATFORMS` the platforms. To publish a new release:

1. Tag the release in the update-apex-ips repository (`git tag -a vX.Y.Z`, push the tag).
2. Put that tag in `TAG` and open a pull request. The build installs exactly that tag and the smoke
   test checks that the image reports it.
3. Merge. The run on `master` publishes and prints the digest to pin.

`smoke-test.sh <image> [expected-version]` can be run by hand against a local build:

```sh
docker buildx build --build-arg VERSION="$(cat TAG)" -t 410labs/update-apex-ips:"$(cat TAG)" --load .
./smoke-test.sh 410labs/update-apex-ips:"$(cat TAG)" "$(cat TAG)"
```
