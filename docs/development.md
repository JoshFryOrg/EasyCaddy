# Development

`docker compose up --build` builds the image from `src/` (after `cp .env.example .env`), then serves
<https://localhost> with a self-signed certificate from this repo's `Caddyfile`.

Dependabot proposes Caddy, module and action updates; merging to `main` publishes the image. To add a module, add a
blank import to `src/tools.go` and run `go mod tidy` in `src/`. The Core Rule Set is pinned by hand in
`src/Dockerfile`; the CRS Watch workflow opens an issue with the new lines when a release is out.

The image is built `FROM scratch`: Caddy, its entrypoint (`src/entrypoint/main.go`, which formats the Caddyfile and
starts Caddy), CA certificates and the WAF's files, and nothing else. There's no shell or package manager, so
`docker exec -it caddy sh` won't work; run Caddy's own commands instead, such as
`docker exec caddy caddy reload --config /etc/caddy/Caddyfile` or `docker run --rm joshfryup/easycaddy caddy version`.

[Back to the README](../README.md)
