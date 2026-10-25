# envy

End-to-end encrypted `.env` sync for developer teams.
Encrypt locally, store ciphertext only, share a project passphrase out-of-band.

## Status

v0.2 — Hosted API: https://envy-api.rovierr.workers.dev. The CLI is open source. Self-host [envy-api](https://github.com/RovierrHQ/envy-api) when you want your own Cloudflare account. `--backend=mock` is for offline tests only.

## Install

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/install.ps1 | iex
```

The script downloads the latest GitHub Release, verifies `checksums.txt`, installs into `~/.local/bin` (Unix) or `%LOCALAPPDATA%\envy\bin` (Windows), and adds that directory to your user PATH. Open a **new** terminal afterward.

Pin a version: `ENVY_VERSION=v0.2.0` (Unix) or `$env:ENVY_VERSION = 'v0.2.0'` (Windows). Skip PATH changes with `--no-path` / `-NoPath`.

Gatekeeper (macOS) or SmartScreen (Windows) may warn on **manual** downloads from Releases — prefer the install script.

**Upgrade** — re-run the install one-liner (overwrites the binary; prints old → new when `envy version` is available).

**Uninstall**

```bash
curl -fsSL https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/uninstall.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/uninstall.ps1 | iex
```

### From source

```bash
go install github.com/RovierrHQ/envy/cmd/envy@latest
```

Or:

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

Owner (hosted API):

```bash
envy login
# Interactive passphrase prompts, or set ENVY_PASSPHRASE for non-interactive use
envy init

printf 'FOO=hello\nBAR=world\n' > .env
envy push --env dev

envy invite create --role write
# share the invite code + project passphrase out-of-band (e.g. 1Password)
```

`envy init` defaults `api_url` to the hosted API and registers a project. Use `--backend=mock` only for offline tests.

Teammate:

```bash
# install envy (see Install), then:
envy login
envy invite accept <code>
# clone the repo (committed .envy.json), then:
envy unlock
envy pull --env dev
```

Store the passphrase out-of-band. If it is lost and no teammate still has the derived key in their OS keyring, remote ciphertext cannot be decrypted.

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
| `envy init` | Create `.envy.json`, derive key, register project on hosted API |
| `envy login` | Device login, or `envy login --token` with a PAT from `/settings/tokens` |
| `envy logout` | Remove the API token from the OS keyring |
| `envy unlock` | Derive project key from passphrase into the OS keyring |
| `envy push --env <name>` | Encrypt local file and upload |
| `envy pull --env <name>` | Download, decrypt, atomic write (use `--force` if dirty) |
| `envy invite create --role read\|write` | Owner mints a 48h single-use invite |
| `envy invite accept <code>` | Join a project (login first) |
| `envy version` | Print CLI version |

Global flag: `--backend=mock|http` (default http when `api_url` is set; mock when empty / `--backend=mock`).

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

Offline mock backend:

```bash
envy --backend=mock init
envy --backend=mock push --env dev
envy --backend=mock pull --env dev
```

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
