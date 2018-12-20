# Host socat tunnel

A simple variation on alpine/socat that port forwards between a docker network
and a host.  Provide the port as the command, or set the environment variable
SOCAT_TUNNEL_PORT. i.e:

```
docker run 410labs/socat-port-fwd-to-host 9100
```
