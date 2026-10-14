# Both stages are pinned to the same readable Caddy version so Dependabot's `docker` updater can bump them. Keep the
# two tags identical: the builder compiles the custom binary that the runtime stage's image then runs.
FROM caddy:2.11.7-builder AS builder

# Module versions come from `go.mod`, which Dependabot's `gomod` updater keeps current (see `tools.go`). Only direct
# requirements are compiled in; `// indirect` lines are their dependencies.
COPY go.mod /tmp/go.mod
RUN set -eu; \
    args=$(awk '/^\t/ && !/\/\/ indirect/ { printf "--with %s@%s ", $1, $2 }' /tmp/go.mod); \
    echo "xcaddy modules: $args"; \
    xcaddy build $args

FROM caddy:2.11.7

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY --chmod=755 custom-entrypoint.sh /custom-entrypoint.sh

ENTRYPOINT ["/custom-entrypoint.sh"]
