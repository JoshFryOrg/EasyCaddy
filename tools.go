//go:build tools

// This file exists only so Dependabot's `gomod` updater can see the Caddy modules compiled into the image. The
// Dockerfile reads the versions out of `go.mod`; nothing here is ever built.
package tools

import (
	_ "github.com/WeidiDeng/caddy-cloudflare-ip"
	_ "github.com/caddy-dns/cloudflare"
	_ "github.com/caddyserver/transform-encoder"
	_ "github.com/fvbommel/caddy-combine-ip-ranges"
	_ "github.com/hslatman/caddy-crowdsec-bouncer"
	_ "github.com/mholt/caddy-l4"
)
