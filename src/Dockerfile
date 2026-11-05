# All three caddy stages are pinned to the same readable Caddy version so Dependabot's `docker` updater can bump them
# together. The builder compiles the custom binary; the stock image is read only for its files.
FROM caddy:2.11.8-builder AS builder

# Module versions come from `go.mod`, which Dependabot's `gomod` updater keeps current (see `tools.go`). Only direct
# requirements are compiled in; `// indirect` lines are their dependencies. The binary is stripped and path-free, like
# the stock one.
COPY go.mod /tmp/go.mod
RUN set -eu; \
    args=$(awk '/^\t/ && !/\/\/ indirect/ { printf "--with %s@%s ", $1, $2 }' /tmp/go.mod); \
    echo "xcaddy modules: $args"; \
    XCADDY_GO_BUILD_FLAGS="-trimpath -ldflags '-s -w'" xcaddy build $args --output /out/caddy; \
    /out/caddy version

# The entrypoint (entrypoint/main.go): the job a shell script did, as the image has no shell. Static, standard
# library only.
COPY entrypoint/main.go /src/entrypoint/main.go
RUN cd /src/entrypoint && CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /out/easycaddy-entrypoint main.go

# The OWASP Core Rule Set for Coraza, a hand pin: no Dependabot ecosystem moves a release tarball, so bump both ARGs
# together (the CRS Watch workflow opens an issue with them). The checksum is of the `-minimal` tarball, so a moved or
# tampered release stops the build. Only the rules, their data files, the setup file and the licence are kept.
ARG CRS_VERSION=4.30.0
ARG CRS_SHA256=3d678a41fd5aade34760127fef5dd64fd7a77848913fc0f70dde0cf467c94427
RUN set -eu; \
    wget -q -O /tmp/crs.tar.gz \
      "https://github.com/coreruleset/coreruleset/releases/download/v${CRS_VERSION}/coreruleset-${CRS_VERSION}-minimal.tar.gz"; \
    echo "${CRS_SHA256}  /tmp/crs.tar.gz" | sha256sum -c -; \
    mkdir -p /tmp/crs /out/coraza/crs; \
    tar -xzf /tmp/crs.tar.gz -C /tmp/crs --strip-components=1; \
    cp -r /tmp/crs/rules /tmp/crs/LICENSE /out/coraza/crs/; \
    cp /tmp/crs/crs-setup.conf.example /out/coraza/crs/crs-setup.conf
COPY --chmod=644 coraza.conf selfhosted.conf strict.conf /out/coraza/

# The directories Caddy uses, empty, made here where there's a shell. /tmp is Coraza's scratch space; /var/log/caddy
# is where a Caddyfile's file logs usually go.
RUN mkdir -p /root-out/data/caddy /root-out/config/caddy /root-out/etc/caddy /root-out/srv /root-out/tmp \
      /root-out/var/log/caddy \
 && chmod 1777 /root-out/tmp

# The stock image, read for its runtime files alone: the default Caddyfile and page, CA certificates and mime types.
FROM caddy:2.11.8 AS stock

# Nothing but Caddy, its entrypoint and their files: no shell, package manager or other programs, so nothing an
# attacker who took over Caddy could use, and nothing for a vulnerability scanner to flag. Caddy is a static Go binary.
FROM scratch
COPY --from=builder /root-out/ /
COPY --from=stock /etc/caddy/Caddyfile /etc/caddy/Caddyfile
COPY --from=stock /usr/share/caddy /usr/share/caddy
COPY --from=stock /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=stock /etc/mime.types /etc/mime.types
COPY --from=builder /out/caddy /usr/bin/caddy
COPY --from=builder /out/easycaddy-entrypoint /usr/bin/easycaddy-entrypoint
# Outside /etc/caddy, so mounting a Caddyfile directory over it never hides the rules.
COPY --from=builder /out/coraza /usr/share/easycaddy/coraza
COPY --chmod=644 waf.caddy /usr/share/easycaddy/waf.caddy

# The stock image's settings.
ENV PATH=/usr/bin \
    XDG_CONFIG_HOME=/config \
    XDG_DATA_HOME=/data
WORKDIR /srv
EXPOSE 80 443 443/udp 2019

ENTRYPOINT ["/usr/bin/easycaddy-entrypoint"]
