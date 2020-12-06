# promtail

This is based on the official [grafana/promtail] image. All it does is add the 
command line argument `-client.external-labes="#{EXTERNAL_LABELS}"` so you can 
set labels via an environment variable.

E.g. in `docker-compose.yml`, used with docker swarm:

```yaml
  promtail:
    image: "410labs/promtail"
    deploy:
      mode: global
    environment:
      EXTERNAL_LABELS: "host={{.Node.Hostname}}"
```
