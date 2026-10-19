# EasyCaddy

[Caddy](https://caddyserver.com) with these modules already built in:

- [Cloudflare DNS](https://github.com/caddy-dns/cloudflare) (`dns.providers.cloudflare`)
- [Cloudflare IP ranges](https://github.com/WeidiDeng/caddy-cloudflare-ip) (`http.ip_sources.cloudflare`)
- [Combine IP ranges](https://github.com/fvbommel/caddy-combine-ip-ranges)
- [Layer 4](https://github.com/mholt/caddy-l4) (TCP/UDP proxying)
- [Transform encoder](https://github.com/caddyserver/transform-encoder) (log formats)
- [CrowdSec bouncer](https://github.com/hslatman/caddy-crowdsec-bouncer) (`crowdsec`, `appsec`)
- [Coraza WAF](https://github.com/corazawaf/coraza-caddy) (`coraza_waf`), with the [OWASP Core Rule Set](https://coreruleset.org) bundled
- [Rate limit](https://github.com/mholt/caddy-ratelimit) (`rate_limit`)
- [OIDC sign-in](https://github.com/relvacode/caddy-oidc) (`oidc`), for Authentik, Keycloak, Entra ID or any OpenID Connect provider

Nothing is downloaded when the container starts, so it boots quickly and works offline.

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

## Writing your Caddyfile

EasyCaddy ships no site config: you bring your own `Caddyfile`. The examples below show how to use the built-in modules.
Replace the domains, addresses and secrets with your own.

### Global options

```caddyfile
{
	email admin@example.com

	# Prometheus metrics at the admin endpoint (/metrics), labelled per site.
	# Keep the admin API off the internet: bind it to localhost or a private address.
	admin 127.0.0.1:2019
	metrics {
		per_host
	}

	servers {
		# Behind Cloudflare: trust its IP ranges and read the real client IP from its header.
		trusted_proxies cloudflare
		client_ip_headers Cf-Connecting-Ip

		# Optional: no time limit on reading request bodies (large, slow uploads).
		timeouts {
			read_body 0
			write 0
		}
	}

	# CrowdSec bouncer, only if you run CrowdSec. appsec_url enables its WAF (the `appsec` directive).
	crowdsec {
		api_url http://crowdsec.internal:8080
		api_key REPLACE_WITH_BOUNCER_KEY
		appsec_url http://crowdsec.internal:7422
		ticker_interval 15s
	}

	# Access log to the container's output...
	log {
		output stdout
	}
	# ...and to a file CrowdSec can read (mount /var/log/caddy as a volume).
	log access {
		output file /var/log/caddy/access.log {
			roll_size 30MiB
			roll_keep 5
		}
	}
}
```

To trust Cloudflare and your own proxies together, use `trusted_proxies combine { cloudflare  static 10.0.0.0/8 }`.
For DNS challenges (wildcards, or a server not reachable on port 80) add `acme_dns cloudflare REPLACE_WITH_API_TOKEN`
to the global options, with a Cloudflare API token that has Zone:DNS:Edit.

### Web application firewall (Coraza)

The WAF is **off unless you turn it on**. The image ships it ready to use: the OWASP Core Rule Set, engine settings
tuned for a reverse proxy, and two named routes, `waf` and `waf_detect`, in `/usr/share/easycaddy/waf.caddy`. To opt
in, import that file once, after the global options:

```caddyfile
import /usr/share/easycaddy/waf.caddy
```

Then put `invoke waf` in the route of each site to protect (see [Sites](#sites)), and set `EASYCADDY_WAF_MODE` to `On` in
the container's environment. Until then `invoke waf` lets everything through, so a Caddyfile can carry it and the WAF
be switched on or off for every site at once. A named route is built once and shared, so fifty
sites cost one copy of the Rule Set in memory. `invoke waf_detect` runs the same rules but only logs what they match,
whatever the mode below: use it to try a new app before blocking.

Configure it from the container's environment, not the Caddyfile:

| Variable | Default | What it does |
| --- | --- | --- |
| `EASYCADDY_WAF_MODE` | `Off` | `On` blocks, `DetectionOnly` logs matches without blocking, `Off` lets everything through |
| `EASYCADDY_WAF_PROFILE` | `selfhosted` | `selfhosted` allows what self-hosted apps need (below); `strict` is the Rule Set's stock policy, for plain websites |
| `EASYCADDY_WAF_PARANOIA` | `1` | Rule Set paranoia level, 1 to 4. Higher catches more and has more false positives |
| `EASYCADDY_WAF_ANOMALY_THRESHOLD` | `5` | Score at which a request is blocked. Raise it (e.g. `10`) to block only clearer attacks |
| `EASYCADDY_WAF_BODY_INSPECT_BYTES` | `1048576` | How much of each request body is inspected; the rest streams on uninspected |
| `EASYCADDY_WAF_DIRECTIVES` | (none) | Extra Coraza directives, one per line, applied after the rules, such as `SecRuleRemoveById 942100` |

```yaml
services:
  caddy:
    image: joshfryup/easycaddy:latest
    environment:
      EASYCADDY_WAF_MODE: "On"
      EASYCADDY_WAF_PARANOIA: "1"
      # Each blocked request is logged with its rule's id ([id "942100"]); switch off one that misfires:
      EASYCADDY_WAF_DIRECTIVES: |
        SecRuleRemoveById 942100
```

Changing a variable takes effect when the container is recreated (`docker compose up -d`).

The `selfhosted` profile allows, on top of the Rule Set's defaults:

- **Methods** PUT, PATCH and DELETE, which registries, Plex, file apps and REST APIs use.
- **Body types** raw binary (`application/octet-stream`, tus uploads), plain text, and Docker and OCI image manifests.
  A raw binary body is left unread, while its method, URL, query and headers are still checked.
- **Headers** `Expect: 100-continue` (sent by curl, .NET, rclone and others before a large upload) and `Content-Range`
  (resumable and chunked uploads). The Rule Set refuses both by default, which fails large uploads.

Use either Coraza or CrowdSec's `appsec` on a site, not both: each inspects every request.

What it costs under heavy traffic:

- **Uploads**: at most `EASYCADDY_WAF_BODY_INSPECT_BYTES` of each request body is held while it's inspected; the rest
  streams to the app as it arrives. A 50 GB upload costs 1 MiB, not 50 GB.
- **Downloads and streams**: responses are never inspected or buffered, so downloads of any size (1 TB and more),
  Range requests, Plex streams, server-sent events and websockets pass straight through.
- **Memory**: about 50 MiB for Caddy with the Rule Set loaded, measured at about 120 MiB peak while moving 2 GB
  uploads and downloads through it.
- **CPU**: the rules run once per request, on its URL, headers and the inspected part of its body. Leave the WAF off
  sites that only serve static files or media segments if CPU is short.

### Sites

`route` runs the handlers in the order written: ban check first, then the WAF, then the app.

```caddyfile
app.example.com, www.example.com {
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.10:3000
	}
}
```

CrowdSec's own AppSec WAF in place of Coraza:

```caddyfile
app2.example.com {
	route {
		crowdsec
		appsec
		reverse_proxy 10.0.0.10:3000
	}
}
```

Large uploads and long downloads. `request_body` caps a body's size (Caddy sets no cap, but one bounds abuse), and the
`timeouts` global option above stops slow transfers being cut off:

```caddyfile
files.example.com {
	request_body {
		max_size 50GB
	}
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.11:3000
	}
}
```

Very large files (hundreds of GB to over 1 TB) need nothing extra from Caddy or the WAF: a download streams through
at constant memory (measured flat at under 80 MiB across a 30 GB download through the WAF, at about 430 MB/s), and an
interrupted download resumes with a Range request. What does matter:

- Keep `write 0` (and `read_body 0` for uploads) in the `timeouts` global option, so a transfer that takes hours isn't
  cut off.
- Don't put `encode` (gzip, zstd) on these sites: compressing already-compressed files costs CPU for nothing, and a
  compressed response can't be resumed with a Range request.
- Behind Cloudflare, use a DNS-only (grey cloud) record for these sites. Cloudflare's proxy caps upload size by plan
  (100 MB on Free) and its terms limit serving large non-web files through it.

If an app's uploads still trip the WAF (some upload protocols send binary in a form field), skip the WAF on those
paths only:

```caddyfile
uploads.example.com {
	@upload path /api/upload /api/upload-*
	route {
		crowdsec
		handle @upload {
			reverse_proxy 10.0.0.11:3000
		}
		handle {
			invoke waf
			reverse_proxy 10.0.0.11:3000
		}
	}
}
```

A Docker registry, WAF on: the `selfhosted` profile lets pushes through (PUT, PATCH, DELETE, layer and manifest types). The
registry needs its own login or Caddy's `basic_auth`, as the Docker CLI can't sign in with OIDC. Hash a password with
`docker run --rm joshfryup/easycaddy caddy hash-password`:

```caddyfile
registry.example.com {
	request_body {
		max_size 20GB
	}
	route {
		crowdsec
		invoke waf
		basic_auth {
			admin $2a$14$REPLACE_WITH_HASH
		}
		reverse_proxy 10.0.0.13:5000
	}
}
```

Plex (or Jellyfin): streams and websockets pass through the WAF untouched. Plex's own apps also reach it over plain
HTTP, hence both schemes:

```caddyfile
http://plex.example.com, https://plex.example.com {
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.14:32400
	}
}
```

Rate limiting, per client IP (`{client_ip}` honours `trusted_proxies`, so behind Cloudflare it is the visitor's IP).
Don't rate-limit sites whose clients make many requests by design (Plex, a registry, a photo library):

```caddyfile
api.example.com {
	rate_limit {
		zone per_client {
			key {client_ip}
			events 100
			window 1m
		}
	}
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.12:8080
	}
}
```

Plain HTTP on an address or port (no certificate), a redirect, and a static reply:

```caddyfile
http://10.0.0.5:8080 {
	respond "Hello from the LAN"
}

old.example.com {
	redir https://new.example.com{uri}
}

:80, example.com {
	respond "Hello, World!"
}
```

### Sign-in (OIDC)

The `oidc` directive signs visitors in at Caddy against any OpenID Connect provider (Authentik, Keycloak, Entra ID,
Google), so an app needs no sign-in of its own and no separate proxy outpost. The provider is set once in the global
options; each site then says who may in.

In Authentik: create an **OAuth2/OpenID Provider** (confidential client) whose redirect URIs include
`https://app.example.com/oauth2/callback` for each site that uses it (a regex such as
`https://.*\.example\.com/oauth2/callback` covers them all), and an **Application** using it. Its issuer is
`https://auth.example.com/application/o/<application-slug>/`.

```caddyfile
{
	oidc {
		issuer https://auth.example.com/application/o/home/
		client_id REPLACE_WITH_CLIENT_ID
		client_secret REPLACE_WITH_CLIENT_SECRET
		scope openid email profile
		username email
		authenticate bearer
		authenticate cookie {
			name easycaddy_session
			# 32 random characters: openssl rand -hex 16
			secret REPLACE_WITH_32_RANDOM_CHARACTERS
			# Copy the groups claim into the session, for `claim groups` rules.
			id_claim groups
			max_age 168h
		}
	}
}
```

The Caddyfile now holds these secrets, so keep it out of version control or any public place. Then, per site:

```caddyfile
# Any signed-in user.
wiki.example.com {
	route {
		crowdsec
		invoke waf
		oidc {
			allow {
				user *
			}
		}
		reverse_proxy 10.0.0.20:3000
	}
}

# Only members of one Authentik group, with a health check left open.
grafana.example.com {
	route {
		crowdsec
		invoke waf
		oidc {
			allow {
				claim groups admins
			}
			allow {
				anonymous
				path /api/health
			}
		}
		reverse_proxy 10.0.0.21:3000
	}
}
```

- A browser that isn't signed in is sent to the provider and back; any other client (an API call, a CLI) gets `401`.
- An `allow` rule with no `user` or `claim` match lets anonymous visitors in. Always name who may in.
- The sign-in is a cookie per site; `domain .example.com` in the `cookie` block shares it across subdomains.
- The username is `{http.auth.user.id}`, for logs, or to pass to an app that trusts such a header with
  `header_up X-Forwarded-User {http.auth.user.id}` inside `reverse_proxy`.
- Keep the provider's own site (`auth.example.com`) outside `oidc`: proxy it normally, or its sign-in page would need a
  sign-in.

A site you want switched off can stay in the file commented out with `#`. Check a file before deploying with
`docker run --rm -v ./Caddyfile:/etc/caddy/Caddyfile joshfryup/easycaddy caddy validate --config /etc/caddy/Caddyfile`.

### CrowdSec allowlists

CrowdSec bans on patterns in Caddy's access log, so some apps' normal traffic gets their users banned (Blazor fetching
`/_framework/` files, `docker push` to a registry). Allowlist it with parsers in CrowdSec's (not EasyCaddy's)
`config/parsers/s02-enrich/` directory, one file each, then restart CrowdSec. These match CrowdSec's own Caddy log
parser's fields (`evt.Meta.*`), so install the `crowdsecurity/caddy` collection first.

`config/parsers/s02-enrich/blazor-whitelist.yaml`, for a Blazor WebAssembly app:

```yaml
name: local/blazor
description: "Allowlist events from Blazor"
filter: "evt.Meta.service == 'http' && evt.Meta.log_type in ['http_access-log', 'http_error-log']"
whitelist:
  reason: "Blazor whitelist"
  expression:
    # Config files Blazor probes for (a 404 is normal)
    - evt.Meta.http_status in ['200', '304', '404'] && evt.Meta.http_verb == 'GET' && evt.Meta.http_path in ['/appsettings.json', '/appsettings.Development.json']
    # The framework's own files
    - evt.Meta.http_status in ['200', '304', '404'] && evt.Meta.http_verb == 'GET' && evt.Meta.http_path startsWith '/_framework/'
    # SignalR hubs
    - evt.Meta.http_status == '200' && evt.Meta.http_verb in ['GET', 'POST'] && evt.Meta.http_path startsWith '/hub'
    # The app's own API, when it succeeds
    - evt.Meta.http_status in ['200', '204'] && evt.Meta.http_verb in ['GET', 'POST', 'PATCH'] && evt.Meta.http_path startsWith '/api'
```

`config/parsers/s02-enrich/registry-whitelist.yaml`, for a Docker registry, locked to its own hostname:

```yaml
name: local/registry
description: "Allowlist Docker Registry API calls"
filter: "evt.Meta.service == 'http'"
whitelist:
  reason: "Docker Registry API whitelist"
  expression:
    - "evt.Parsed.host == 'registry.example.com' && evt.Meta.http_path matches '^/v[23]/.*'"
```

`config/parsers/s02-enrich/ip-whitelist.yaml`, so your own address (or network) is never banned:

```yaml
name: local/ip-whitelist
description: "Never ban my own addresses"
whitelist:
  reason: "Trusted addresses"
  ip:
    - "203.0.113.10"
  cidr:
    - "192.168.0.0/16"
```

Allowlists only stop CrowdSec's log-based bans. A request the `appsec` or Coraza WAF blocks is still blocked; skip the
WAF for those paths in the Caddyfile instead (see the uploads examples above). Check a parser loaded with
`cscli parsers list`, and test one against a log line with `cscli explain`.

## Local development

`docker-compose.yml` in this repo builds the image from source instead of pulling it:

```bash
cp .env.example .env
docker compose up --build
```

Then open <https://localhost> (the certificate is self-signed). Edit `Caddyfile` to try your own config.

## Maintaining

Dependabot opens PRs for new Caddy releases, module updates and workflow actions. Each PR is test-built, and merging it to `main` publishes a new image. To get notified, watch this repo's pull requests on GitHub, and choose *Watch > Custom > Releases* on [caddyserver/caddy](https://github.com/caddyserver/caddy).

To add a module, add a blank import to `src/tools.go` and run `go mod tidy` in `src/`.

The OWASP Core Rule Set is pinned by hand (`CRS_VERSION` and `CRS_SHA256` in `src/Dockerfile`), as Dependabot can't
move a release tarball. The CRS Watch workflow checks the pin daily and, when a newer release is out, opens an issue
with the two replacement lines, checksum included.
