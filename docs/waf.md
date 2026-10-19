# Web application firewall

Off until you opt in, in two places:

1. **Caddyfile**: `import /usr/share/easycaddy/waf.caddy` once, then `invoke waf` in each site to protect.
2. **`.env`**: set `EASYCADDY_WAF_MODE=On`, then `docker compose up -d`.

Every setting is explained in [`.env.example`](../.env.example). `invoke waf_detect` in a site only logs what the rules
match, whatever the mode, for trying a new app. Use Coraza or CrowdSec's `appsec` on a site, not both.

Built for heavy traffic: the rules are loaded once for all sites (about 50 MiB), only the first 1 MiB of a request body
is inspected while the rest streams on, and responses are never buffered. A 30 GB download through it held Caddy under
80 MiB; registry pushes, Plex and websockets pass.

[Back to the README](../README.md)
