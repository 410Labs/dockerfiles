# update-apex-ips: keeps a zone apex A record in step with a subdomain's IPs (Route53).
# Source: https://github.com/410Labs/update-apex-ips (MIT). Published as 410labs/update-apex-ips by
# .github/workflows/publish.yml; see ../README.md for how images here are built and published.
#
# The binary is the tagged module version (TAG, passed as the VERSION build argument), fetched
# through the Go module proxy and checked against the checksum database by `go install`, then
# cross-compiled for the target. The build stage always runs on the build host, so the arm64 image
# builds without emulation.
#
# To bump: tag a new release in the update-apex-ips repository, put that tag in ./TAG, and merge.
# The base images are pinned by digest; to refresh them:
#   docker buildx imagetools inspect golang:1.26
#   docker buildx imagetools inspect gcr.io/distroless/static:nonroot

ARG GO_IMAGE=golang:1.26@sha256:cfa27959837144bf4649802ba38e59daffd0dd2a7994aa83daaf113285b0f551
ARG RUNTIME_IMAGE=gcr.io/distroless/static:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
# No default: the build fails if VERSION is not given, rather than building something unnamed.
ARG VERSION
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
# go install puts a cross-compiled binary in /go/bin/<os>_<arch>/, a native one in /go/bin/.
RUN test -n "$VERSION" \
 && GOOS=$TARGETOS GOARCH=$TARGETARCH go install -trimpath "github.com/410Labs/update-apex-ips@$VERSION" \
 && mkdir /out \
 && cp "$(find /go/bin -type f -name update-apex-ips)" /out/update-apex-ips

FROM ${RUNTIME_IMAGE}
COPY --from=build /out/update-apex-ips /update-apex-ips
ENTRYPOINT ["/update-apex-ips"]
