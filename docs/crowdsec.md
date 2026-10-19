# CrowdSec allowlists

Some apps' normal traffic gets their users banned by CrowdSec. Allowlist it with a parser in CrowdSec's
`config/parsers/s02-enrich/` (needs the `crowdsecurity/caddy` collection), then restart CrowdSec. Examples:

```yaml
# blazor-whitelist.yaml: Blazor WebAssembly's own files and calls
name: local/blazor
description: "Allowlist Blazor"
filter: "evt.Meta.service == 'http' && evt.Meta.log_type in ['http_access-log', 'http_error-log']"
whitelist:
  reason: "Blazor"
  expression:
    - evt.Meta.http_status in ['200', '304', '404'] && evt.Meta.http_verb == 'GET' && evt.Meta.http_path in ['/appsettings.json', '/appsettings.Development.json']
    - evt.Meta.http_status in ['200', '304', '404'] && evt.Meta.http_verb == 'GET' && evt.Meta.http_path startsWith '/_framework/'
    - evt.Meta.http_status == '200' && evt.Meta.http_verb in ['GET', 'POST'] && evt.Meta.http_path startsWith '/hub'
    - evt.Meta.http_status in ['200', '204'] && evt.Meta.http_verb in ['GET', 'POST', 'PATCH'] && evt.Meta.http_path startsWith '/api'
```

```yaml
# registry-whitelist.yaml: a Docker registry's API, on its own hostname only
name: local/registry
description: "Allowlist Docker Registry API"
filter: "evt.Meta.service == 'http'"
whitelist:
  reason: "Docker Registry"
  expression:
    - "evt.Parsed.host == 'registry.example.com' && evt.Meta.http_path matches '^/v[23]/.*'"
```

```yaml
# ip-whitelist.yaml: never ban your own addresses
name: local/ip-whitelist
description: "Trusted addresses"
whitelist:
  reason: "Trusted addresses"
  ip:
    - "203.0.113.10"
  cidr:
    - "192.168.0.0/16"
```

Allowlists stop bans only; a WAF block still blocks. Check with `cscli parsers list` and `cscli explain`.

[Back to the README](../README.md)
