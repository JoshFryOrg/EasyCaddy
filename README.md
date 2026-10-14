# EasyCaddy

[Caddy](https://caddyserver.com) with these modules already built in:

- [Cloudflare DNS](https://github.com/caddy-dns/cloudflare) (`dns.providers.cloudflare`)
- [Cloudflare IP ranges](https://github.com/WeidiDeng/caddy-cloudflare-ip) (`http.ip_sources.cloudflare`)
- [Combine IP ranges](https://github.com/fvbommel/caddy-combine-ip-ranges)
- [Layer 4](https://github.com/mholt/caddy-l4) (TCP/UDP proxying)
- [Transform encoder](https://github.com/caddyserver/transform-encoder) (log formats)
- [CrowdSec bouncer](https://github.com/hslatman/caddy-crowdsec-bouncer)

Nothing is downloaded when the container starts, so it boots quickly and works offline.

## Run it

```yaml
services:
  caddy:
    image: docker.joshfry.uk/easycaddy:latest
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

Put your site config in `./Caddyfile`. On start the container tidies its formatting (if the file is writable) and then runs Caddy.

To update, pull the latest image and recreate the container:

```bash
docker compose pull && docker compose up -d
```

## Local development

`docker-compose.yml` in this repo builds the image from source instead of pulling it:

```bash
cp .env.example .env
docker compose up --build
```

Then open <https://localhost> (the certificate is self-signed). Edit `Caddyfile` to try your own config, and set `CLOUDFLARE_API_TOKEN` in `.env` if it uses Cloudflare DNS.

## Maintaining

Dependabot opens PRs for new Caddy releases, module updates and workflow actions. Each PR is test-built, and merging it to `main` publishes a new image. To get notified, watch this repo's pull requests on GitHub, and choose *Watch > Custom > Releases* on [caddyserver/caddy](https://github.com/caddyserver/caddy).

To add a module, add a blank import to `tools.go` and run `go mod tidy`.
