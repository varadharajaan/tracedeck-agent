# Collection Policy

TraceDeck uses typed YAML policy rather than free-form collection flags.

Example:

```yaml
collection:
  browser:
    url_mode: domain_only
    collect_page_title: false
  foreground_app:
    enabled: true
    window_title_mode: none
  sensitive_capabilities:
    credentials: deny
    keystrokes: deny
    cookies: deny
    tokens: deny
    screenshots: deny
```

Policy code lives under:

```text
agent/internal/config/
agent/internal/constants/
docs/schema/policy-v1alpha1.schema.json
```

Schema changes should be generated from Go types, not pasted as string blobs.
