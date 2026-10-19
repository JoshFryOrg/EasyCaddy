# Sign-in (OIDC)

Signs visitors in at Caddy, so apps need no sign-in or proxy outpost of their own. In Authentik, create an
**OAuth2/OpenID Provider** (confidential) with redirect URI `https://<site>/oauth2/callback` (a regex like
`https://.*\.example\.com/oauth2/callback` covers every site) and an **Application** using it; its issuer is
`https://auth.example.com/application/o/<slug>/`. Then fill in the `oidc` global option (as in [caddyfile.md](caddyfile.md)), and in each site:

- `allow { user * }` lets in anyone signed in; `allow { claim groups admins }` only that group. Always name who may in:
  an `allow` with neither lets anonymous visitors in.
- Browsers are sent to sign in; other clients get `401`.
- Don't put `oidc` on the provider's own site.

[Back to the README](../README.md)
