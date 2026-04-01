# Stockyard Cipher

**Secrets manager for CI/CD.** Store environment variables encrypted, expose them to pipelines via short-lived token API. The thing every team needs before they're big enough for Vault. Single binary, AES-256-GCM encryption at rest.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
# Generate an encryption key
export CIPHER_ENCRYPTION_KEY=$(openssl rand -hex 32)

curl -sfL https://stockyard.dev/install/cipher | sh
cipher
```

## Usage

```bash
# 1. Create a project
curl -X POST http://localhost:8870/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"my-api-prod"}'

# 2. Store secrets (encrypted at rest)
curl -X PUT http://localhost:8870/api/projects/{id}/secrets/DATABASE_URL \
  -H "Content-Type: application/json" \
  -d '{"value":"postgres://user:pass@db:5432/myapp"}'

curl -X PUT http://localhost:8870/api/projects/{id}/secrets/API_KEY \
  -H "Content-Type: application/json" \
  -d '{"value":"sk_live_abc123"}'

# 3. Create a short-lived token for your CI pipeline
curl -X POST http://localhost:8870/api/projects/{id}/tokens \
  -H "Content-Type: application/json" \
  -d '{"name":"github-actions","ttl_minutes":15}'
# → Returns raw_token (save it, shown only once)

# 4. In your CI pipeline, fetch secrets with the token
curl -H "Authorization: Bearer {token}" http://localhost:8870/api/secrets

# Or as .env format for direct sourcing
eval $(curl -sH "Authorization: Bearer {token}" "http://localhost:8870/api/secrets?format=env")
```

## Free vs Pro

| Feature | Free | Pro ($2.99/mo) |
|---------|------|----------------|
| Projects | 2 | Unlimited |
| Secrets | 20 | Unlimited |
| Tokens | 5 | Unlimited |
| AES-256-GCM encryption | ✓ | ✓ |
| Short-lived tokens | ✓ | ✓ |
| Audit log | — | ✓ |
| Version history | — | ✓ |

## License

Apache 2.0 — see [LICENSE](LICENSE).
