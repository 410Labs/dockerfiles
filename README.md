# dockerfiles
Public Docker images

Each directory holds one image, published to Docker Hub as `410labs/<directory>`. There are two
kinds:

- **Published by CI:** directories with a `TAG` file, currently `update-apex-ips` and `openbao`.
  See [Images published by CI](#images-published-by-ci).
- **Legacy:** everything else (`graphite`, `statsd`, `promtail`, ...). These were built by hand or
  by Docker Hub automated builds. The automated builds were retired in November 2026, and no
  webhook from this repository triggers a build any more. A change to a legacy directory publishes
  nothing. To change one of these images, move it to CI first (see [Moving a legacy image to
  CI](#moving-a-legacy-image-to-ci)).

## Images published by CI

A directory that has these files is built, tested, and published by
[`.github/workflows/publish.yml`](.github/workflows/publish.yml):

| File | Purpose |
| --- | --- |
| `Dockerfile` | The image. The workflow passes the tag as the `VERSION` build argument. |
| `TAG` | One line: the tag to publish, exactly (no `latest`). |
| `smoke-test.sh` | Executable. `smoke-test.sh <image> <tag>` exits non-zero if the image is wrong. |
| `PLATFORMS` | Optional. Comma-separated, `linux/amd64` (default) and/or `linux/arm64`. |
| `README.md` | What it is, how to run it, how to bump it. |

What the workflow does:

- **Pull requests and other branches** build the image for `linux/amd64` and run its smoke test.
  Nothing is pushed and no credential is available.
- **A run on `master`** also publishes, from the `publish` job only, and only if
  `410labs/<directory>:<TAG>` does not already exist. A published tag is never overwritten: a run
  whose tag exists warns and skips, so to publish new content you change `TAG`. The repository on
  Docker Hub must already exist and be public (a first push could create it private); the run stops
  with a clear message if it is not.
- After a push, the smoke test runs again on the image as published, by digest.
- The run summary prints `410labs/<directory>:<tag>@<digest>`. Consumers pin that, not the tag.

A directory without a `TAG` file is never published. The workflow can also be run by hand
(`workflow_dispatch`, input `image`); it publishes only when run on `master`.

### Adding an image

1. **Create the Docker Hub repository** `410labs/<name>` in the Docker Hub web UI, and make it
   public. The Docker CLI can't create a repository, and a first push would create a private one.
   Don't push a placeholder tag to create it.
2. **Give the CI token Read & Write on the new repository** (a repository admin; see the settings
   below). Without this, the publish job fails with an authorization error.
3. **Add the directory** with the files listed above, and open a pull request. CI builds the image
   and runs the smoke test; nothing is pushed.
4. **Merge.** The run on `master` publishes `<TAG>`, then smoke-tests the image again by digest.
5. **Pin the result** where the image is used. Take `410labs/<name>:<tag>@<digest>` from the run
   summary and use the tag and the digest together, never the tag alone.

### Releasing a new version

1. In a pull request, change `TAG` and whatever the Dockerfile installs. Use a new tag every time:
   a tag that already exists is skipped, not overwritten.
2. Merge, take the new `tag@digest` from the run summary, and update each consumer's pin in its
   own repository.

To withdraw a bad release, pin consumers back to the previous `tag@digest`. Published tags stay on
Docker Hub, so the previous digest is always available.

### Moving a legacy image to CI

Add `TAG` (a tag not yet used on Docker Hub), an executable `smoke-test.sh` and a `README.md` to
its directory, check that the Dockerfile builds from this repository alone, and open a pull
request. From the merge on, the image follows the release process above. Existing tags on Docker
Hub are left as they are.

### Settings the workflow relies on (repository admins)

- Environment `publish`, limited to the `master` branch, holding the variable `DOCKER_USERNAME` and
  the secret `DOCKER_PASSWORD`. That must be a Docker Hub access token, not an account password, with
  Read & Write only on the `410labs` repositories published from here. A new CI image needs the token
  granted Read & Write on its repository. When the token is rotated, update the secret and confirm
  that the next publish succeeds.
- A ruleset on `master` requiring pull requests, and approval before workflows run for fork
  pull requests. Do not add a `pull_request_target` trigger.
- No webhooks to Docker Hub. Publishing happens only through this workflow.
- Actions are pinned by commit SHA. Bump them in a pull request, preferring releases that are at
  least two weeks old.
