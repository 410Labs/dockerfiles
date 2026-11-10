# update-apex-ips

Keeps a zone apex A record in step with a subdomain. A zone apex can't be a CNAME, so when
`www.example.com` points at a load balancer, the apex `example.com` needs the same IPs as plain A records.

```
update-apex-ips -name www.example.com           # plan only
update-apex-ips -name www.example.com -apply    # update Route53
```

1. Resolves `-name` (IPv4) with the system resolver.
2. Looks up the apex (derived with the public suffix list) the same way. If it already matches, stops.
3. Otherwise finds the public Route53 hosted zone (or use `-zone-id`), reads the apex A record, and
   prints the difference. With `-apply`, UPSERTs the apex to the source's IPs.

Exit codes: `0` in sync or applied, `1` error, `2` drift found in plan-only mode.

An existing record keeps its TTL; a new record gets 300 s. It refuses to run when the name is already the
apex, when the source has no IPv4 addresses, or when the apex record is an alias or a weighted, latency or
failover set.

## AWS access

Standard AWS SDK credential chain. The principal needs `route53:ListHostedZones` (account-wide),
and `route53:ListResourceRecordSets` and `route53:ChangeResourceRecordSets` (scopeable to one zone).

## The image

`410labs/update-apex-ips:<tag>` on Docker Hub: public, `linux/amd64` and `linux/arm64`, distroless, runs as
a non-root user. The tag is the contents of [`TAG`](TAG), and `-version` prints it. Tags are never
overwritten and there is no `latest`: pin `<tag>@<digest>` (the digest is in the publish run's summary).

```sh
docker run --rm -e AWS_REGION=us-east-1 \
  -v "$PWD/credentials:/aws/credentials:ro" -e AWS_SHARED_CREDENTIALS_FILE=/aws/credentials \
  410labs/update-apex-ips:v0.1.0 -name www.example.com
```

CI (`.github/workflows/publish.yml`, see the repository README) builds and publishes it. The
[`Dockerfile`](Dockerfile) builds from the source in this directory and fails unless the code is
gofmt-clean and passes `go vet` and `go test`.

To release a change: edit the code and [`TAG`](TAG) together in one pull request. The pull request builds
and smoke-tests the image; the run on `master` after the merge publishes it.

```sh
# the same build and test, by hand
docker buildx build --build-arg VERSION="$(cat TAG)" -t 410labs/update-apex-ips:"$(cat TAG)" --load .
./smoke-test.sh 410labs/update-apex-ips:"$(cat TAG)" "$(cat TAG)"
```

## Running in a pod

The program uses the standard AWS credential chain, so the pod only needs a credentials file. Use
short-lived STS credentials from the OpenBao/Vault AWS secrets engine (`credential_type=assumed_role`);
they work immediately, with no IAM propagation delay. Two ways to deliver them:

| Delivery | Manifests | Use when |
|---|---|---|
| External Secrets Operator (`VaultDynamicSecret`) | [`deploy/eso/`](deploy/eso) | ESO is already installed |
| Vault agent injector | [`deploy/injector/`](deploy/injector) | the injector is installed |

### One-time setup (both options)

**AWS** (an administrator; the policy documents are in [`deploy/iam/`](deploy/iam), with placeholders):

1. Create the IAM user `ENGINE_USER` with [`engine-user-policy.json`](deploy/iam/engine-user-policy.json) (it may only
   `sts:AssumeRole` the role below, and create, list and delete its own access keys so OpenBao can run `bao write -f aws/config/rotate-root`). Create no access key until the OpenBao engine is ready to receive it.
2. Create the role `update-apex-ips` with trust policy [`trust-policy.json`](deploy/iam/trust-policy.json) and permissions
   [`permissions-policy.json`](deploy/iam/permissions-policy.json). The permissions allow only UPSERT of A records, only on the
   apex names, in the listed zones; the default 1 h maximum session is enough.
3. Later, create one access key for `ENGINE_USER` and write it straight into OpenBao as the engine's root
   (`bao write aws/config/root`). Keep it out of shell history, logs, and chat.

**OpenBao/Vault** (keep keys out of shell history):

```
bao secrets enable aws
bao write aws/config/root region=us-east-1 access_key=... secret_key=...
bao write aws/roles/update-apex-ips \
    credential_type=assumed_role \
    role_arns=arn:aws:iam::ACCOUNT_ID:role/update-apex-ips \
    default_sts_ttl=1h max_sts_ttl=1h
bao policy write update-apex-ips - <<'POLICY'
path "aws/creds/update-apex-ips" { capabilities = ["read"] }
POLICY
bao write auth/AUTH_MOUNT/role/update-apex-ips \
    bound_service_account_names=update-apex-ips-eso \
    bound_service_account_namespaces=NAMESPACE \
    policies=update-apex-ips ttl=15m
```

OpenBao does not bundle the AWS secrets engine in its core binary. If `bao secrets enable aws` reports an
unknown plugin, declare the `secrets-aws` plugin from
[openbao-plugins](https://github.com/openbao/openbao-plugins) in the server configuration
(`plugin "secret" "aws" { ... }` with its image, version and `sha256sum`) and restart the servers.

### Cluster

Apply `deploy/eso/credentials.yaml` and `deploy/eso/cronjob.yaml` after replacing the placeholders. For the
injector, apply `deploy/injector/cronjob.yaml` instead (its init-only mode,
`vault.hashicorp.com/agent-pre-populate-only`, lets the Job finish).

Start plan-only: leave `-apply` off, so a drift exits 2 and the Job shows as failed, which is a useful
alert while you watch the first runs. Trigger one run by hand to prove the credential chain:
`kubectl -n NAMESPACE create job --from=cronjob/update-apex-ips first-run`.

Notes:
- The credentials template reads `security_token`. Check the field names with `bao read aws/creds/update-apex-ips`.
- The public-DNS check goes through the cluster DNS (CoreDNS). A rewrite or split-horizon rule for the name inside the cluster can hide drift.

## Caveats

- The public check uses the system resolver, so a stale cache can hide drift or report a false mismatch.
- Copying IPs only makes sense if the source's A records are stable. A source that rotates IPs or CNAMEs to a CDN will make the apex chase it.

## Build and test

```
go build -o update-apex-ips .
go test ./...
```

## License

MIT; see the repository's [LICENSE](../LICENSE).
