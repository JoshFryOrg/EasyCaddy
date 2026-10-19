# Writing your Caddyfile

One example covering every module. Keep what you need, and replace the names, addresses and `REPLACE_WITH_` values.
The file then holds secrets, so keep it private.

```caddyfile
{
	email admin@example.com

	servers {
		# Behind Cloudflare: the real client IP comes from its header.
		trusted_proxies cloudflare
		client_ip_headers Cf-Connecting-Ip
		# No time limit on transfers, for large uploads and downloads.
		timeouts {
			read_body 0
			write 0
		}
	}

	# Only if you run CrowdSec. appsec_url enables its `appsec` WAF.
	crowdsec {
		api_url http://crowdsec.internal:8080
		api_key REPLACE_WITH_BOUNCER_KEY
		appsec_url http://crowdsec.internal:7422
	}

	# Only for sign-in at Caddy (see docs/oidc.md).
	oidc {
		issuer https://auth.example.com/application/o/home/
		client_id REPLACE_WITH_CLIENT_ID
		client_secret REPLACE_WITH_CLIENT_SECRET
		scope openid email profile
		username email
		authenticate bearer
		authenticate cookie {
			name easycaddy_session
			secret REPLACE_WITH_32_RANDOM_CHARACTERS # openssl rand -hex 16
			id_claim groups
			max_age 168h
		}
	}

	log {
		output stdout
	}
}

# The WAF (see docs/waf.md). Blocks nothing until you turn it on in .env.
import /usr/share/easycaddy/waf.caddy

# A typical site: ban check, then the WAF, then the app. `route` keeps that order.
app.example.com, www.example.com {
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.10:3000
	}
}

# Large uploads and downloads: cap the body size.
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

# A Docker registry. The Docker CLI can't use OIDC, so basic auth
# (hash: docker run --rm joshfryup/easycaddy caddy hash-password).
registry.example.com {
	request_body {
		max_size 20GB
	}
	route {
		crowdsec
		invoke waf
		basic_auth {
			admin REPLACE_WITH_HASH
		}
		reverse_proxy 10.0.0.13:5000
	}
}

# Plex: its apps also use plain HTTP.
http://plex.example.com, https://plex.example.com {
	route {
		crowdsec
		invoke waf
		reverse_proxy 10.0.0.14:32400
	}
}

# Signed in through OIDC, members of one group only.
grafana.example.com {
	route {
		crowdsec
		invoke waf
		oidc {
			allow {
				claim groups admins
			}
		}
		reverse_proxy 10.0.0.21:3000
	}
}

# Rate limited per client: 100 requests a minute.
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

# Skip the WAF on one path, for an app whose uploads trip it.
uploads.example.com {
	@upload path /api/upload*
	route {
		crowdsec
		handle @upload {
			reverse_proxy 10.0.0.15:3000
		}
		handle {
			invoke waf
			reverse_proxy 10.0.0.15:3000
		}
	}
}

# Plain HTTP on a LAN address, and a redirect.
http://10.0.0.5:8080 {
	respond "Hello from the LAN"
}

old.example.com {
	redir https://new.example.com{uri}
}
```

Check a file before deploying:
`docker run --rm -v ./Caddyfile:/etc/caddy/Caddyfile joshfryup/easycaddy caddy validate --config /etc/caddy/Caddyfile`.

For DNS challenges (wildcard certificates) add `acme_dns cloudflare REPLACE_WITH_API_TOKEN` to the global options, with
a token that has Zone:DNS:Edit.

[Back to the README](../README.md)
