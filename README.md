# EasyCaddy

Docker image of [Caddy](https://caddyserver.com) with the modules compiled in at build time, pushed to `docker.joshfry.uk/easycaddy:latest`.

## Modules

- `github.com/caddy-dns/cloudflare`
- `github.com/WeidiDeng/caddy-cloudflare-ip`
- `github.com/fvbommel/caddy-combine-ip-ranges`
- `github.com/mholt/caddy-l4`
- `github.com/caddyserver/transform-encoder`
- `github.com/hslatman/caddy-crowdsec-bouncer`

## Startup

The entrypoint formats `/etc/caddy/Caddyfile` (best effort, so a read-only mount is fine) and then runs Caddy. Nothing is downloaded at start, so the container no longer needs network access to boot.

## Keeping up to date

Dependabot checks daily and opens PRs for:

- **Caddy**: the `caddy:<version>-builder` and `caddy:<version>` tags in the `Dockerfile`, bumped together.
- **Modules**: `go.mod`, whose direct requirements the `Dockerfile` passes to `xcaddy` as `--with module@version`. `tools.go` exists only so the modules appear as requirements.
- **Workflow actions**.

Merging any of them to `main` runs `.github/workflows/publish.yml`, which rebuilds and pushes the image from a self-hosted runner (needs the `DOCKER_PASSWORD` secret).

To add a module, add a blank import to `tools.go` and run `go mod tidy`.
