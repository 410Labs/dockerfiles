# StatsD docker image

Compared to the official statsd Dockerfile:
 * Runs as non-priviledged `statsd` user.
 * Uses node 8.9.4.
 * Patches statsd 0.8.0 to work with recent versions of node.  See:
    https://github.com/etsy/statsd/compare/02eae13...8d5363c

If you are happy with the basic config, and graphite/carbon is running on a host
named "graphite" (e.g. using `--add-host` or `--link` or docker's embedded DNS
to match service names), you can just run it without additional configuration:

    # standalone
    docker run -d --name statsd \
      --add-host graphite:10.0.0.123 \
      410Labs/statsd:0.8.0

    # or in a swarm, which is running a 'graphite' service
    docker service create --name statsd 410Labs/statsd:0.8.0

If you need to customize the config file, override `/etc/statsd/config.js`:

    docker run -d --name statsd \
      -v /path/to/your/config:/etc/statsd/config.js \
      410Labs/statsd:0.8.0
