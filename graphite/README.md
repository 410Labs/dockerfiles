# Graphite web docker image

Similar to the official graphite image, but:
 * squashes most of the `RUN` layers
 * removes nginx and statsd to their own images
 * separate graphite, carbon-cache, and carbon-aggregator into different
   containers (using the same base image)
 * config files are not volumes (by default; you are free to override at
   runtime)
 * TODO: rebase on a non-phusion base image...

### Running
```
docker volume create graphite
docker run --name graphite -v graphite:/opt/graphite/storage 410labs/graphite
```

You will want to also run an nginx proxy (see [410labs/nginx-graphite]) to use
the web UI, and a `carbon-cache` container for inserting data (see
[410labs/carbon-cache]).  Run it on the same server as `carbon-cache`, sharing
the same storage volume. The nginx proxy can live on a different server.

[410labs/nginx-graphite]: https://hub.docker.com/r/410labs/nginx-graphite/
[410labs/carbon-cache]: https://hub.docker.com/r/410labs/carbon-cache/

# Customizing

You can override individual config files in `/opt/graphite/conf` using bind
mounts or swarm configs. Or you could mount a volume for the entire
`/opt/graphite/conf` directory.  You should use the same config files for
`graphite` as you use for `carbon-cache`.
