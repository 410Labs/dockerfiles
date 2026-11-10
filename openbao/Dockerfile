# OpenBao with the AWS KMS seal plugin and the AWS secrets engine plugin baked in.
# Published as 410labs/openbao by .github/workflows/publish.yml; see README.md.
#
# OpenBao 2.7 serves the AWS KMS seal only through the external plugin
# openbao-plugin-kms-aws. With plugin_auto_download, every start downloads it
# from ghcr into the plugin directory, so a ghcr outage blocks every bao
# restart. This image puts the binary where the downloader would, under the
# name it would use (<type>-<name>-<version>, here kms-awskms-v0.1.1). At
# start OpenBao finds it, checks it against the config's sha256sum, and logs
# "plugin is cached on disk, skipping download". If the file were missing or
# wrong it would fall back to downloading, so the config keeps its plugin
# stanza (image, version, sha256sum) unchanged.
#
# The repository is public: the image holds only upstream open-source binaries, so it can be
# pulled without credentials, for example while the secrets store itself is down.
#
# The secrets-aws plugin (the AWS secrets engine) is baked for a related reason. The OCI
# downloader leaves a plugin as a symlink into .oci-cache, and registering that by -command is
# refused ("cannot execute files outside of configured plugin directory"); a regular file here
# registers. (Registering from an OCI image reference may be allowed; not tested.) Its file name
# is <type>-<name>-<version>: type `secret`, name `aws`. OpenBao does not auto-register it. Once
# per cluster, with the binary sha256 from the ARG below:
#   bao plugin register -sha256=<binary sha256> -version=<v> -command=secret-aws-<v> secret aws
#   bao secrets enable -plugin-version=<v> aws
#
# To bump a plugin or OpenBao:
#  1. Change the ARGs below. Base digest from `docker buildx imagetools inspect
#     ghcr.io/openbao/openbao:<version>`. Checksums from the plugin release's
#     checksums-<plugin>.txt, but first verify the release tarball's .sig against OpenBao's GPG
#     key (fingerprint 66D15FDD87287219C8E15478D200CD702853E6D0, from
#     https://openbao.org/assets/openbao-gpg-pub-20240618.asc) and take hashes only from a
#     tarball that verifies. The release itself is mutable and its hashes are unsigned.
#  2. Put the new tag in ./TAG, and update `plugins` (file names and binary sha256) in
#     ./smoke-test.sh. That list repeats the ARGs on purpose, so the smoke test checks the built
#     image independently of the Dockerfile's own checks.
#  3. In each consumer: the image tag and digest, and where a plugin is declared, its sha256sum.
#  4. A registered secrets plugin names the old file (-command=secret-aws-<old>), and the mount
#     is pinned to -plugin-version. After rolling to an image without the old file, credential
#     minting fails until you register the new version, run `bao secrets tune
#     -plugin-version=<new> aws`, then `bao plugin reload -mounts=aws`. Keep the old file in
#     the image for one release to avoid the gap or to roll back.

ARG OPENBAO_IMAGE=ghcr.io/openbao/openbao:2.7.2@sha256:5046911075e137e154ce49ab051502f4ec7cf367d484ae4a4dce665d9b90561f

FROM alpine:3.20 AS plugin
ARG KMS_AWS_VERSION=v0.1.1
# openbao-plugin-kms-aws_linux_amd64_v1.tar.gz, from checksums-kms-aws.txt.
ARG KMS_AWS_TARBALL_SHA256=f00ebed7fa1cbec080e0fed908e167125bc14b266df7e4af845f82aa4650ed4e
# The binary inside it; the same value as sha256sum in the bao configs.
ARG KMS_AWS_BINARY_SHA256=0dffe0a159a5bdbeb147a892595fe517de858f48a99a64f33fabc775af0eec76
ARG SECRETS_AWS_VERSION=v0.3.2
# openbao-plugin-secrets-aws_linux_amd64_v1.tar.gz of release secrets-aws-v0.3.2: the asset
# digest GitHub shows for it. The .sig of this tarball verifies against OpenBao's key
# (see "To bump" above).
ARG SECRETS_AWS_TARBALL_SHA256=5563b5490135732ec68dda46643d81f5eb28714d1137652c34bd69f7e9a55ce4
# The binary inside it, from checksums-secrets-aws.txt: the value for `bao plugin register
# -sha256` and for the sha256sum in the plugin stanza.
ARG SECRETS_AWS_BINARY_SHA256=c828ea1fd34a9d2c2c5546346a659216b871a5774a1beaa306061161f68a0a30
RUN apk add --no-cache curl \
 && curl -fsSLo /tmp/kms-aws.tar.gz \
      "https://github.com/openbao/openbao-plugins/releases/download/kms-aws-${KMS_AWS_VERSION}/openbao-plugin-kms-aws_linux_amd64_v1.tar.gz" \
 && echo "${KMS_AWS_TARBALL_SHA256}  /tmp/kms-aws.tar.gz" | sha256sum -c - \
 && mkdir -p /out \
 && tar -xzf /tmp/kms-aws.tar.gz -C /tmp openbao-plugin-kms-aws_linux_amd64_v1 \
 && install -m 0755 /tmp/openbao-plugin-kms-aws_linux_amd64_v1 "/out/kms-awskms-${KMS_AWS_VERSION}" \
 && echo "${KMS_AWS_BINARY_SHA256}  /out/kms-awskms-${KMS_AWS_VERSION}" | sha256sum -c - \
 && curl -fsSLo /tmp/secrets-aws.tar.gz \
      "https://github.com/openbao/openbao-plugins/releases/download/secrets-aws-${SECRETS_AWS_VERSION}/openbao-plugin-secrets-aws_linux_amd64_v1.tar.gz" \
 && echo "${SECRETS_AWS_TARBALL_SHA256}  /tmp/secrets-aws.tar.gz" | sha256sum -c - \
 && tar -xzf /tmp/secrets-aws.tar.gz -C /tmp openbao-plugin-secrets-aws_linux_amd64_v1 \
 && install -m 0755 /tmp/openbao-plugin-secrets-aws_linux_amd64_v1 "/out/secret-aws-${SECRETS_AWS_VERSION}" \
 && echo "${SECRETS_AWS_BINARY_SHA256}  /out/secret-aws-${SECRETS_AWS_VERSION}" | sha256sum -c -

FROM ${OPENBAO_IMAGE}
# CI passes the tag (the contents of ./TAG) as VERSION.
ARG VERSION
LABEL org.opencontainers.image.title="openbao" \
      org.opencontainers.image.description="OpenBao with the AWS KMS seal and AWS secrets engine plugins baked in" \
      org.opencontainers.image.source="https://github.com/410Labs/dockerfiles" \
      org.opencontainers.image.version="${VERSION}"
# The directory stays writable by openbao, so the download fallback still works.
COPY --from=plugin --chown=openbao:openbao /out/ /openbao/plugins/
