# envy

End-to-end encrypted `.env` sync for developer teams.
Encrypt locally, store ciphertext only, share a project passphrase out-of-band.

## Status

v0 — CLI talks to [envy-api](https://github.com/RovierrHQ/envy-api) (`https://envy-api.rovierr.workers.dev` once deployed). `--backend=mock` stays for offline tests. Custom domain `envy.rovierr.com` waits until that zone is on the Cloudflare account.

## Install

```bash
go install github.com/RovierrHQ/envy/cmd/envy@latest
```

Or build from source:

```bash
git clone https://github.com/RovierrHQ/envy.git
cd envy
go build -o bin/envy ./cmd/envy
```

Cross-compile release binaries:

```bash
make cross
```

## Quick start

```bash
# Interactive passphrase prompts, or set ENVY_PASSPHRASE for non-interactive use
envy --backend=mock init

printf 'FOO=hello\nBAR=world\n' > .env
envy --backend=mock push --env dev

rm .env
envy --backend=mock pull --env dev
cat .env
```

Teammates clone the repo (which includes committed `.envy.json`), run `envy unlock` with the shared passphrase, then `envy pull`.

## How it works

- **Key derivation:** Argon2id over the shared passphrase and a salt stored in `.envy.json` produces a 32-byte key. The key is kept in the OS credential store; the passphrase is not.
- **Verifier:** A BLAKE2b-256 of the derived key lives in `.envy.json` so a wrong passphrase fails locally before any network call.
- **Envelope:** Ciphertext is `envy1 || nonce || AES-256-GCM(ciphertext+tag)`, base64-encoded in the API payload.
- **Zero-knowledge server:** Only `{ "ciphertext", "version" }` is stored. Salt and KDF parameters never leave the repo config.
- **Environments:** Named envs (`dev`, `staging`, `prod`) map to local files via `.envy.json` `files`. Optimistic concurrency rejects stale pushes.

`.envy.json` is committed. `.envy/state.json` (local version + content hash) is not.

## Commands

| Command | Purpose |
| --- | --- |
| `envy init` | Create `.envy.json`, derive key, optionally register a project |
| `envy login` | Device login, or `envy login --token` with a PAT from `/settings/tokens` |
| `envy logout` | Remove the API token from the OS keyring |
| `envy unlock` | Derive project key from passphrase into the OS keyring |
| `envy push --env <name>` | Encrypt local file and upload |
| `envy pull --env <name>` | Download, decrypt, atomic write (use `--force` if dirty) |
| `envy invite create --role read\|write` | Owner mints a 48h single-use invite |
| `envy invite accept <code>` | Join a project (login first) |

Global flag: `--backend=mock|http` (default mock when `api_url` is empty).

## API contract

```
POST /v1/projects
PUT  /v1/projects/{id}/secrets/{env}   body: { ciphertext, version }  # version = client's base
GET  /v1/projects/{id}/secrets/{env}   → { ciphertext, version }
POST /v1/projects/{id}/invites         body: { role }
POST /v1/invites/accept                body: { code }
POST /v1/device/code
POST /v1/device/token
```

`Authorization: Bearer <token>`. `409` on version conflict.

## Development

```bash
go test ./...
go build -o bin/envy ./cmd/envy
```

```
cmd/envy/          entrypoint
internal/cli/      cobra commands
internal/config/   .envy.json + local state
internal/crypto/   Argon2id, verifier, envy1 seal/open
internal/api/      StorageBackend, HTTP + mock
internal/keyring/  OS credential store
internal/fsutil/   atomic writes
```

## License

MIT
