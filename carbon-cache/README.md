# Graphite carbon-cache image

Based on [410labs/graphite] image, this image simply changes the exposed ports
and the command to `carbon-cache` instead of `graphite-web`.  Run it on the same
server as graphite, using a shared volume.

### Running

```
docker volume create graphite
docker run --name carbon-cache -v graphite:/opt/graphite/storage 410labs/carbon-cache
docker run --name graphite     -v graphite:/opt/graphite/storage 410labs/graphite
docker run --name nginx-graphite --link graphite 410labs/nginx-graphite
```

# Customizing

You can override individual config files in `/opt/graphite/conf`
using bind mounts or swarm configs. Or you could mount a volume for
the entire `/opt/graphite/conf` directory.

[410labs/graphite]: https://hub.docker.com/r/410labs/graphite/
