# StatsD docker image

Compared to the official statsd Dockerfile:
 * Runs as non-priviledged `statsd` user.
 * Uses node 8.9.4.
 * Patches statsd 0.8.0 to work with recent versions of node.  See:
    https://github.com/etsy/statsd/compare/02eae13...8d5363c

### Running

If you are happy with the standard statsd configuration, and carbon-cache is
available on a host named "carbon-cache" (e.g. using `--add-host` or `--link` or
docker's embedded DNS to match service names), you can just run it without much
additional configuration.

Standalone container:
`docker run -d --name statsd --add-host carbon-cache:10.0.0.123 410Labs/statsd:0.8.0`

In a swarm, which is running a 'carbon-cache' service:
`docker service create --name statsd 410Labs/statsd:0.8.0`

#### Custom config file

To customize the config file, override `/etc/statsd/config.js`, using volumes,
bind mount, or swarm configs/secrets. E.g:
```
docker run -d --name statsd \
  -v /path/to/your/config:/etc/statsd/config.js \
  410Labs/statsd:0.8.0
```
