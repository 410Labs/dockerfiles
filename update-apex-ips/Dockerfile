# update-apex-ips: keeps a zone apex A record in step with a subdomain's IPs (Route53).
# Published as 410labs/update-apex-ips by .github/workflows/publish.yml; see README.md.
#
# Built from the source in this directory. The build stage fails unless the code is gofmt-clean and
# passes `go vet` and `go test`, so a published image has been tested. Dependencies are checked
# against go.sum. The binary reports VERSION (the contents of ./TAG) from `-version`. The build
# stage runs on the build host and cross-compiles, so the arm64 image needs no emulation.
#
# To bump: change the code and ./TAG together in one pull request. The base images are pinned by
# digest; to refresh them:
#   docker buildx imagetools inspect golang:1.26
#   docker buildx imagetools inspect gcr.io/distroless/static:nonroot

ARG GO_IMAGE=golang:1.26@sha256:cfa27959837144bf4649802ba38e59daffd0dd2a7994aa83daaf113285b0f551
ARG RUNTIME_IMAGE=gcr.io/distroless/static:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
# No default: the build fails if VERSION is not given, rather than building something unnamed.
ARG VERSION
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN set -eu; \
    [ -n "${VERSION:-}" ] || { echo "the VERSION build argument is required" >&2; exit 1; }; \
    unformatted=$(gofmt -l .); \
    [ -z "$unformatted" ] || { echo "gofmt needed on: $unformatted" >&2; exit 1; }; \
    go vet ./...; \
    go test ./...; \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -trimpath -ldflags "-X main.buildVersion=$VERSION" -o /out/update-apex-ips .

FROM ${RUNTIME_IMAGE}
COPY --from=build /out/update-apex-ips /update-apex-ips
ENTRYPOINT ["/update-apex-ips"]
