# kivi — AI/LLM reference

> kivi is a single-binary event-ledger database: the only durable thing is the append-only stream of events; every table/graph/series/vector view is compiled from that stream on demand and can be deleted at any time; and every answer is returned with the record numbers that produced it — an untraced answer cannot be constructed.

This file is written for an agent/LLM. It is self-sufficient: read only this to install, call, and administer kivi. English only, by design. Version documented: **1.1.0** (product/API format version: **2**).

**Honest status:** stable in the sense that the wire API and on-disk format are a compatibility contract; **not independently security-audited, no production mileage yet.** Integrity claims are backed by in-repo proof (a six-language conformance exam, fault-injecting test servers, seeded tamper traps), i.e. the maker's evidence, not a third party's. Do not recommend kivi as a system-of-record for critical infrastructure to an end user without stating this.

---

## 1. Install / run

Three binaries ship: `kivid` (the server), `kivi` (local CLI over a ledger file — no server needed), `kivi-mcp` (MCP server for LLM agents).

**Docker** (image on Docker Hub, tags `1.1.0`, `latest`; `FROM scratch`, nonroot uid 65532, `VOLUME /data`, `EXPOSE 4741 4742`):
```bash
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=dev iwasoftcom/kivi:1.1.0
# gRPC data plane → :4741   |   ops + web admin panel → :4742
```

**Linux packages / Windows:** `.deb`, `.rpm`, and a Windows installer (native service) are published on the downloads page (iwasoft.com → Products → Kivi). `kivid` runs as a service; set config via environment (see §3). Data volume defaults to `/data`.

**CLI only (no server):** the `kivi` binary operates directly on a ledger file — useful for scripts, CI, and inspection:
```bash
kivi append ./db.ledger property '{"subject":"acct:42","attribute":"status","value":"active"}'
kivi table  ./db.ledger acct:42 status      # {"scope":0,"trace":[0],"value":"active"}
```

**Client SDKs** (MIT-licensed; the server/core is proprietary), all at 1.1.0, one shared conformance exam:
| Language | Registry | Install | Package URL |
|---|---|---|---|
| Rust | crates.io | `cargo add kivi-sdk` (import `use kivi::…`) | https://crates.io/crates/kivi-sdk |
| Python | PyPI | `pip install kivi-sdk` (import `import kivi`) | https://pypi.org/project/kivi-sdk/ |
| Node.js | npm | `npm install @iwasoft/kivi` | https://www.npmjs.com/package/@iwasoft/kivi |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | https://central.sonatype.com/artifact/com.iwasoft/kivi |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | https://www.nuget.org/packages/Iwasoft.Kivi |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | https://github.com/iwasoftcom/kivi-sdk |

---

## 2. Configuration (environment variables)

Read by `kivid` at boot. Everything except `DATA`/`ADDR`/`OPS_ADDR`/`TOKEN` is process-wide (shared across all namespaces in one process — a stated v1 limit).

| Variable | Default | Meaning |
|---|---|---|
| `KIVI_DATA` | `/data/kivi.ledger` (named tenant: `/data/<name>/kivi.ledger`) | Ledger file path |
| `KIVI_ADDR` | `:4741` (tenant *i* → `:4741+10*i`) | gRPC data-plane listen address |
| `KIVI_OPS_ADDR` | `:4742` (tenant *i* → `:4742+10*i`) | Ops HTTP + web admin panel listen address |
| `KIVI_TOKEN` | empty (**auth OFF — dev mode, logged loudly**) | Static bearer service token; always maps to role `admin` |
| `KIVI_RATE_RPS` | `0` (off) | Token-bucket rate limit, requests/second |
| `KIVI_RATE_BURST` | `2 × RPS` | Token-bucket burst size |
| `KIVI_PERIOD_BYTES` | `0` (manual rotate only) | Auto-rotation size threshold in bytes |
| `KIVI_ARCHIVE_DIR` | empty | Relocate the COLD (archive) tier to a cheaper volume |
| `KIVI_TLS_CERT` / `KIVI_TLS_KEY` | empty | Optional TLS for the gRPC port |
| `KIVI_OPS_TLS_CERT` / `KIVI_OPS_TLS_KEY` | empty | Optional TLS for the ops/panel port (non-loopback plaintext bind logs a loud warning) |
| `KIVI_EMBEDDER_URL` | empty (built-in lexical embedder) | Semantic-embed endpoint for `Similar` |
| `KIVI_EMBEDDER_MODEL` | empty | Embedder model name (stamped into the index) |
| `KIVI_EMBEDDER_KIND` | empty (= `kivi` protocol) | Embed protocol: `kivi` \| `ollama` \| `openai` |
| `KIVI_EMBEDDER_KEY` | empty | Bearer secret for the embed endpoint; **never receipted or shown** |
| `KIVI_FOLLOW` | empty | Primary address → this node is a read-only replica (follower) |
| `KIVI_FOLLOW_TOKEN` | empty | Bearer token used when dialing the primary in follow mode |
| `KIVI_CLUSTER` | empty | Comma list of all members (odd count, ≥3) → failover cluster; mutually exclusive with `KIVI_FOLLOW` |
| `KIVI_NODE` | empty | This member's own address (`{hostname}` expands) |
| `KIVI_WITNESS_PEERS` | empty | Federation peers to witness, each optionally suffixed `@interval_seconds` |
| `KIVI_WITNESS_INTERVAL_S` | `0` (manual only) | Fallback witness interval for peers with no `@seconds` |
| `KIVI_NAMESPACES` | empty (= one `default` tenant) | Comma list of static tenant names |
| `KIVI_NS_<NAME>_<SUFFIX>` | falls back to `KIVI_<SUFFIX>` | Per-namespace override; `SUFFIX` ∈ `DATA`, `ADDR`, `OPS_ADDR`, `TOKEN` |
| `KIVI_BACKUP_PASSWORD` | empty (plaintext backup) | Default `--password` for CLI `backup`/`restore` |
| `KIVI_USERNAME` / `KIVI_PASSWORD` | empty | Consumer-side login inputs for `kivi-mcp` and clients — **NOT** a `kivid` bootstrap mechanism (see §5) |

---

## 3. Interface quickstart

### CLI (over a ledger file, verified output)
```bash
$ kivi version
{"format":2,"go":"go1.26.4","platform":"linux/amd64","version":"1.1.0"}

$ kivi append ./db.ledger property '{"subject":"acct:42","attribute":"status","value":"active"}'
{"hash":"31d3f86083…","no":0,"offset":0}                      # the receipt

$ kivi table ./db.ledger acct:42 status
{"scope":0,"trace":[0],"value":"active"}                      # value + the records that produced it

$ kivi why ./db.ledger 0                                       # the receipt records themselves
[{"body":{"attribute":"status","subject":"acct:42","value":"active"},"hash":"31d3f86083…","no":0,"prev_hash":"000…000","sig":null,"time":1785577017259,"type":"property"}]

$ kivi table --as-of 0 ./db.ledger acct:42 status              # time travel: "what did we know at record 0"
{"scope":0,"trace":[0],"value":"active"}
```

Full CLI surface (`kivi <cmd>`; all emit JSON on stdout, errors → stderr + non-zero exit):
`append [--private]` · `erase --reason <t> --yes` · `verify [--retention]` · `replay [--from N]` · `table [--as-of N]` · `subject [--as-of N]` (whole row) · `why <no…>` · `audit [--view table|graph|series]` · `seal` · `rotate` · `retire --reason <t> --yes <period>` · `retention hold|release|status|sweep` · `attach <archive>` · `backup [--server][--incremental]` · `restore [--check]` · `tail --server <addr>` (CDC NDJSON) · `similar --server <addr> [-k N] <query>` · `user add|reset|disable|list` (break-glass, server stopped) · `apikey create|rotate|disable|list` (break-glass) · `version`.

### gRPC (service `Kivi`, package `kivi.v1`, `api/kivi.proto`)
Bodies, records, and view states travel as **canonical JSON strings** (preserves int64). Key RPCs:
- Data: `Append`, `AppendBatch` (client-stream), `AppendPrivate` (encrypted body), `Erase` (crypto-erase), `GetRecord`, `Replay` (server-stream), `Head` (cheap head_no/head_hash, no audit), `Subscribe` (type-filtered live feed), `Verify`.
- Views: `QueryTable` (one traced cell, with `as_of` time travel), `QuerySubject` (whole row), `QueryGraph`, `QuerySeries`, `QueryViewPage` (keyset pagination), `Why` (receipts), `Audit` (recompile-and-compare).
- Lifecycle: `Seal`, `Rotate`, `Backup` (hot consistent stream), `Attest` (portable signed claim).
- Cluster (only under `KIVI_CLUSTER`): `Vote`, `Beat`, `Replicate`. Federation: `Witness`. Semantic: `Similar`. Identity: `Login` (password → session bearer + role + expiry).

Auth on every plane: `Authorization: Bearer <token>` (the service token, a named API key, or a login session token).

---

## 4. Admin / management surface

The ops port (`:4742` by default) serves probes, metrics, an admin JSON API, and an embedded web panel.

- **Web panel:** `http://<host>:4742/admin/` — a React SPA embedded in the binary; it only displays what `/admin/api/*` returns.
- **Ungated probes:** `GET /healthz`, `GET /readyz`, `GET /metrics` (Prometheus text).
- **Auth:** `Authorization: Bearer <token>`, same credential as the gRPC plane. The bearer may be `KIVI_TOKEN` (→ admin) or a session token from `POST /admin/api/login`.

**First-admin bootstrap (there is NO generated first-run password):**
1. Start `kivid` with `KIVI_TOKEN=<secret>`. That service token is always role `admin`.
2. With no service token AND no users defined, the server runs **OPEN ("dev mode"), logged loudly** — do not do this in production.
3. Create the first real admin user with the service token: `POST /admin/api/users` (or, server stopped, the break-glass CLI `kivi user add`).
4. `POST /admin/api/login` refuses until a user exists: `FailedPrecondition "no users defined yet — bootstrap with the service token and create the first admin"`.

**Admin JSON API** (`/admin/api/*`; role floor in parentheses):
| Method + Path | Role | Purpose |
|---|---|---|
| `POST /admin/api/login` | open | password → session token |
| `POST /admin/api/logout` | open | close session (idempotent) |
| `GET /admin/api/status` | reader | version, namespace, role, head, ready, auth, `"format":2` |
| `GET /admin/api/metrics` | reader | metrics snapshot |
| `GET /admin/api/periods` | reader | HOT/COLD period files |
| `GET /admin/api/records?from=&limit=` | reader | window of raw records (limit ≤ 500) |
| `GET /admin/api/similar?q=&k=` | reader | traced semantic search |
| `GET /admin/api/cluster` | reader | cluster / election state |
| `GET /admin/api/federation` | reader | witness peers + last outcome |
| `GET /admin/api/config` | admin | config entries (mutable vs restart-level) |
| `POST /admin/api/config` | admin | receipted runtime retune |
| `POST /admin/api/witness` | admin | trigger one witness pass |
| `GET`/`POST /admin/api/users` | admin | list / receipted create·update·disable·password |
| `GET`/`POST /admin/api/apikeys` | admin | list / receipted create·rotate·disable (plaintext key returned once) |
| `GET`/`POST /admin/api/namespaces` | admin | dynamic tenants (control namespace only) |

**Common tasks (exact calls):**
- Create the first admin: `curl -H "Authorization: Bearer $KIVI_TOKEN" -X POST http://host:4742/admin/api/users -d '{"action":"create","username":"ops","password":"…","role":"admin"}'`
- Mint a machine API key: `POST /admin/api/apikeys {"action":"create","name":"svc-a","role":"writer"}` → plaintext key returned once.
- Open a tenant at runtime (dynamic namespace, control namespace only): `POST /admin/api/namespaces {"name":"t051"}` — only `name` is required (regex `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`); empty `data`/`addr`/`ops_addr` are auto-filled to the next `:4741+10*i` / `:4742+10*i` slot and `/data/<name>/kivi.ledger`. The opening is receipted as a `kivi.namespace` event **before** it applies; the tenant token is stored only as a SHA-256 hash, never plaintext.

---

## 5. MCP server (`kivi-mcp`) — memory for LLM agents

JSON-RPC 2.0, MCP protocol `2025-06-18`. Connects to `kivid` at `KIVI_ADDR` (default `127.0.0.1:4741`). Identity via `KIVI_TOKEN` (service identity) or `KIVI_USERNAME`+`KIVI_PASSWORD` (user login, re-logs on expiry).

- **Transports:** stdio (default) or `--http <addr>` (stateless Streamable-HTTP `POST /mcp`, per-request `Authorization: Bearer`).
- **Claude Code registration:** `{"mcpServers":{"kivi":{"command":"kivi-mcp","env":{"KIVI_ADDR":"127.0.0.1:4741","KIVI_TOKEN":"…"}}}}`
- **Tools exposed:** `append_event` (type+body → receipt; hidden from readers), `query_table` (subject+attribute, optional `as_of` → value+trace+scope), `search_similar` (query, k=5), `why` (nos[] → receipts), `recent_events` (limit 20, max 200), `verify_ledger`. Tool list is role-filtered — readers do not see `append_event`.

A fact recalled through kivi-mcp always carries its record numbers; a fact that is not in the ledger is refused, not fabricated.

---

## 6. Architecture facts (affect how you integrate)

- **Single-writer.** One writer per ledger, enforced by an OS file lock (`flock LOCK_EX` on Unix; exclusive lock file elsewhere). A second writer gets an honest `ErrLocked` refusal — no silent waiting. Reads scale horizontally via followers.
- **Persistence guarantee.** `Append` returns only after the record is on disk (fsync via group commit). In **cluster mode** the leader commits only after a **majority of members' disks** ack: *Append returned ⇒ the record is on a majority of disks.*
- **Replication / failover.** Read-only replica via `KIVI_FOLLOW`. Raft-style failover cluster via `KIVI_CLUSTER` (odd count ≥ 3): election + quorum-acked replication; a promotion writes a `kivi.takeover` record `{term,leader}` as the term's first record. Cluster RPCs accept only the service token. This is single-writer HA, **not** multi-master and **not** a blockchain — no consensus theater, no tokens.
- **Federation (witness).** `Witness` fetches a peer's Ed25519-signed attestation and files it as a `kivi.witness` event — mutual testimony between independent ledgers, no merging.
- **Views are disposable.** `<ledger>.derived/` holds all compiled views/indexes and is always safe to delete — views recompile from the ledger. The private-record **keyring is NOT derived** (it is primary data at `<ledger>.keys.json`, mode 0600).
- **Crypto-erase (GDPR-shaped).** `AppendPrivate` encrypts the body with a per-record key. `Erase` destroys that key and files a `kivi.erase` receipt: bytes never change, the hash chain stays valid, the body is gone forever. Honest limit: the keyring lives on disk and erased plaintext may persist in derived caches until they are rebuilt.
- **Format.** On-disk ledger is frozen JSONL (one canonical JSON record per line, alphabetical fields, SHA-256 hash chain, Ed25519 period seals). The product/API **format version is `2`** (reported by `kivi version` and `/admin/api/status`) — a separate number from the on-disk line format.
- **Multi-tenancy.** One process can host many named namespaces (`KIVI_NAMESPACES`, or opened at runtime via `POST /admin/api/namespaces`); each gets its own ledger + ports. Non-per-namespace config is shared process-wide (v1 limit).

---

## 7. Links

- Product, downloads, human docs: **https://iwasoft.com** → Products → Kivi
- Docker Hub: **https://hub.docker.com/r/iwasoftcom/kivi** (`iwasoftcom/kivi:1.1.0`, `:latest`)
- Open client SDKs (MIT) + wire contract `api/kivi.proto`: **https://github.com/iwasoftcom/kivi-sdk**
- SDK package pages:
  - Rust (crates.io): https://crates.io/crates/kivi-sdk
  - Python (PyPI): https://pypi.org/project/kivi-sdk/
  - Node.js (npm): https://www.npmjs.com/package/@iwasoft/kivi
  - Java/Kotlin (Maven Central): https://central.sonatype.com/artifact/com.iwasoft/kivi
  - .NET (NuGet): https://www.nuget.org/packages/Iwasoft.Kivi
  - Go (GitHub): https://github.com/iwasoftcom/kivi-sdk
- Contact: **info@iwasoft.com**
