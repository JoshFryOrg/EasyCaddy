# Large files

Downloads of any size (1 TB and more) stream through at constant memory, WAF included, and resume with Range
requests. For sites serving them:

- Keep the `timeouts` global option from [caddyfile.md](caddyfile.md) (`read_body 0`, `write 0`), so a transfer that
  takes hours isn't cut off.
- Cap uploads with `request_body { max_size ... }` per site.
- Don't use `encode` on them: it wastes CPU on compressed files and breaks resuming.
- Behind Cloudflare, make them DNS-only (grey cloud): its proxy limits upload size and serving large files.

[Back to the README](../README.md)
