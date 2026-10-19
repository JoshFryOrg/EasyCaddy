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

# The OWASP Core Rule Set for Coraza, a hand pin: no Dependabot ecosystem moves a release tarball, so bump both ARGs
# together from https://github.com/coreruleset/coreruleset/releases (the checksum is of the `-minimal` tarball). A moved
# or tampered release stops the build. Only the rules, their data files, the setup file and the licence are kept.
ARG CRS_VERSION=4.30.0
ARG CRS_SHA256=3d678a41fd5aade34760127fef5dd64fd7a77848913fc0f70dde0cf467c94427
RUN set -eu;     wget -q -O /tmp/crs.tar.gz       "https://github.com/coreruleset/coreruleset/releases/download/v${CRS_VERSION}/coreruleset-${CRS_VERSION}-minimal.tar.gz";     echo "${CRS_SHA256}  /tmp/crs.tar.gz" | sha256sum -c -;     mkdir -p /tmp/crs /out/coraza/crs;     tar -xzf /tmp/crs.tar.gz -C /tmp/crs --strip-components=1;     cp -r /tmp/crs/rules /tmp/crs/LICENSE /out/coraza/crs/;     cp /tmp/crs/crs-setup.conf.example /out/coraza/crs/crs-setup.conf
COPY --chmod=644 coraza.conf selfhosted.conf strict.conf /out/coraza/

FROM caddy:2.11.7

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
# Outside /etc/caddy, so mounting a Caddyfile directory over it never hides the rules.
COPY --from=builder /out/coraza /usr/share/easycaddy/coraza
COPY --chmod=644 waf.caddy /usr/share/easycaddy/waf.caddy
COPY --chmod=755 custom-entrypoint.sh /custom-entrypoint.sh

ENTRYPOINT ["/custom-entrypoint.sh"]
