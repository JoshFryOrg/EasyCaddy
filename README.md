# EasyCaddy

[Caddy](https://caddyserver.com) with these modules built in, so nothing is downloaded at start:

- [Cloudflare DNS](https://github.com/caddy-dns/cloudflare) and [IP ranges](https://github.com/WeidiDeng/caddy-cloudflare-ip), [combined IP ranges](https://github.com/fvbommel/caddy-combine-ip-ranges)
- [CrowdSec bouncer](https://github.com/hslatman/caddy-crowdsec-bouncer) (`crowdsec`, `appsec`)
- [Coraza WAF](https://github.com/corazawaf/coraza-caddy) with the [OWASP Core Rule Set](https://coreruleset.org), opt-in
- [OIDC sign-in](https://github.com/relvacode/caddy-oidc) (`oidc`): Authentik, Keycloak, Entra ID or any OpenID Connect provider
- [Rate limit](https://github.com/mholt/caddy-ratelimit), [Layer 4](https://github.com/mholt/caddy-l4), [transform encoder](https://github.com/caddyserver/transform-encoder)

## Run it

```yaml
services:
  caddy:
    image: joshfryup/easycaddy:latest
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"
    env_file: .env # the WAF's settings; copy .env.example
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

Bring your own `Caddyfile` (below) and a `.env` copied from [`.env.example`](.env.example). To update:
`docker compose pull && docker compose up -d`.

## Your Caddyfile

EasyCaddy ships no site config. A minimal one:

```caddyfile
app.example.com {
	reverse_proxy 10.0.0.10:3000
}
```

## Docs

- [Writing your Caddyfile](docs/caddyfile.md): one example using every module (CrowdSec, WAF, OIDC, rate limits, a
  registry, Plex)
- [Web application firewall](docs/waf.md): opting in, and what it costs under heavy traffic; its settings are explained
  in [`.env.example`](.env.example)
- [Sign-in (OIDC)](docs/oidc.md): signing visitors in at Caddy, with Authentik
- [Large files](docs/large-files.md): uploads and downloads of 1 TB and more
- [CrowdSec allowlists](docs/crowdsec.md): stopping apps' normal traffic getting users banned
- [Development](docs/development.md): building locally and maintaining the image
