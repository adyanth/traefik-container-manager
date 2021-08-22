# traefik-container-manager

Traefik plugin to start/stop containers as needed.

Needs `traefik-container-manager-service` and should be accessible by traefik container to work. Defaults to `http://manager:10000/api`, so if you have a compose file with the service named manager accessible by traefik over the default network, you are good to go.

Just add this middleware to any router, configuring name which should match `traefik-container-manager.name`. and timeout. with the needed labels for `traefik-container-manager-service`.

A sample shown below can be used for reference:

```yaml
  whoami:
    image: containous/whoami
    labels: 
      - traefik.enable=true
      - traefik.http.routers.whoami.entrypoints=entryhttp
      - traefik.http.routers.whoami.rule=Host(`whoami.adyanth.lan`) || PathPrefix(`/whoami`)
      - traefik.http.routers.whoami.middlewares=whoami-timeout
      - traefik.http.services.whoami.loadbalancer.server.port=80
      - traefik.http.middlewares.whoami-timeout.plugin.traefik-container-manager.timeout=5
      - traefik.http.middlewares.whoami-timeout.plugin.traefik-container-manager.name=whoami
      - traefik.http.middlewares.whoami-timeout.plugin.traefik-container-manager.serviceUrl=http://manager:10000/api     # Optional
      - traefik-container-manager.name=whoami
      - traefik-container-manager.path=/whoami # Prefix matched to the incoming path. To be used if using PathPrefix routing
      - traefik-container-manager.host=whoami # Prefix matched to the incoming hostname. Can provide the subdomain or the complete fqdn. To be used if using Host routing
```
