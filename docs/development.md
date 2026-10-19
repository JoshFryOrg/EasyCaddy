# Development

`docker compose up --build` builds the image from `src/` (after `cp .env.example .env`), then serves
<https://localhost> with a self-signed certificate from this repo's `Caddyfile`.

Dependabot proposes Caddy, module and action updates; merging to `main` publishes the image. To add a module, add a
blank import to `src/tools.go` and run `go mod tidy` in `src/`. The Core Rule Set is pinned by hand in
`src/Dockerfile`; the CRS Watch workflow opens an issue with the new lines when a release is out.

[Back to the README](../README.md)
