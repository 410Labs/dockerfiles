# dockerfiles
Public Docker images

Each directory holds one image, published to Docker Hub as `410labs/<directory>`. Most of them
(`graphite`, `statsd`, `promtail`, ...) were built and pushed by hand, and CI leaves them alone.

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

To add an image, create the Docker Hub repository as public, add the directory with the files
above, and open a pull request. To release a new version, change `TAG` (and whatever the
Dockerfile installs) in a pull request. A directory without a `TAG` file is never published.

### Settings the workflow relies on (repository admins)

- Environment `publish`, limited to the `master` branch, holding the variable `DOCKER_USERNAME` and
  the secret `DOCKER_PASSWORD`: a Docker Hub token limited to the `410labs` repositories published
  from here. A new CI image needs the token granted Read & Write on its repository.
- A ruleset on `master` requiring pull requests, and approval before workflows run for fork
  pull requests. Do not add a `pull_request_target` trigger.
- Actions are pinned by commit SHA. Bump them in a pull request, preferring releases that are at
  least two weeks old.
