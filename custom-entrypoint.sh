#!/bin/sh

# A read-only or missing Caddyfile must not stop Caddy starting, so formatting is best-effort.
echo "Formatting Caddyfile"
caddy fmt --overwrite /etc/caddy/Caddyfile || echo "Caddyfile not formatted, continuing"

echo "Launching Caddy"
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile 2>&1
