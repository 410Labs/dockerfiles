# openbao

[OpenBao](https://openbao.org) 2.7 with two plugins baked into the image, so a restart never has to
download them, and one helper:

- the AWS KMS seal plugin (`openbao-plugin-kms-aws`), which OpenBao 2.7 needs for an AWS KMS auto-unseal,
- the AWS secrets engine plugin (`openbao-plugin-secrets-aws`),
- since `2.7.2-aws-kms-sec-ira`, AWS's IAM Roles Anywhere credential helper (`aws_signing_helper` 1.8.6),
  so the KMS seal can use short-lived credentials from an X.509 certificate instead of a static IAM
  access key (410Labs/infrastructure#445).

`410labs/openbao:<tag>` on Docker Hub: public, `linux/amd64`. The tag starts with the OpenBao version and
names what is baked in: `2.7.2-kms-aws-v0.1.1-secrets-aws-v0.3.2` (both plugins), then
`2.7.2-aws-kms-sec-ira` (the same plugins plus the Roles Anywhere helper); the exact versions are in the
`Dockerfile` ARGs. Tags are never overwritten and there is no `latest`: pin
`<tag>@<digest>` (the digest is in the publish run's summary).

Images published before this repository built them were published as `mailstrom/openbao`; the same
manifests (same digests) are available under both names:
`2.7.2-kms-aws-v0.1.1` (KMS plugin only) and `2.7.2-kms-aws-v0.1.1-secrets-aws-v0.3.2`.

## What is in it

The upstream `ghcr.io/openbao/openbao` image (pinned by digest) plus, in `/openbao/plugins/`, regular files
named `<type>-<name>-<version>`: `kms-awskms-v0.1.1` and `secret-aws-v0.3.2`, each checked against its
release checksums at build time and again by `smoke-test.sh`. At start OpenBao finds them, checks them
against the `sha256sum` in its configuration, and logs `plugin is cached on disk, skipping download`.

`/usr/local/bin/aws_signing_helper` is owned by root and not writable by `openbao`, checked against its
pinned sha256 at build time and again by `smoke-test.sh`, which also runs `aws_signing_helper version`.
It is AWS's glibc build and runs through the base image's `gcompat`. To use it for the seal, give the
container an AWS config file whose `[default]` profile has
`credential_process = /usr/local/bin/aws_signing_helper credential-process --certificate ... --private-key ...
--trust-anchor-arn ... --profile-arn ... --role-arn ...`, set `AWS_CONFIG_FILE` to it, and remove any static
`AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` (they take precedence). The image does nothing with it on its
own, so an instance that keeps static keys behaves exactly as on the previous tag.

Keep the `plugin` stanzas in the server configuration (image, version, `sha256sum`): they make a missing or
wrong file fall back to downloading, and the KMS plugin is declared there at all.

The secrets plugin needs registering once per cluster; OpenBao does not do it by itself:

```sh
bao plugin register -sha256=<binary sha256> -version=v0.3.2 -command=secret-aws-v0.3.2 secret aws
bao secrets enable -plugin-version=v0.3.2 aws
```

## Building and bumping

CI builds and publishes it (`.github/workflows/publish.yml`, see the repository README). The header of the
[`Dockerfile`](Dockerfile) has the bump procedure: verify the plugin release's signature, change the ARGs,
`TAG` and the list in `smoke-test.sh` together in one pull request, then update each consumer's tag and
digest. Rolling a consumer onto a new plugin version also needs the register, tune and reload steps in
the Dockerfile header.

```sh
docker buildx build --platform linux/amd64 --build-arg VERSION="$(cat TAG)" -t 410labs/openbao:"$(cat TAG)" --load .
./smoke-test.sh 410labs/openbao:"$(cat TAG)" "$(cat TAG)"
```
