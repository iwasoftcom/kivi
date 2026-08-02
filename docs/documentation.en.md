Kivi

the database that shows its receipts · v1.1.0 · format v2

An event-ledger database in pure Go — one static binary that can account for every answer it gives.

## What Kivi is

Kivi stores only **events** — immutable, hash-chained records. Everything that looks like current state (a table, a graph, a time series, a vector index) is a **view**: a fold over those events, recompiled on demand and disposable at any moment. Every answer carries a **trace** (the record numbers that established it) and a **scope** (how far into history it looked), and `why` returns the receipt records themselves. Missing data is an honest refusal, never a fabrication.

## The three principles

**Only events are durable.**  
There is no UPDATE. New facts are appended; the past is never rewritten. History is first-class.

**Representations are compiled.**  
Deleting `derived/` is always legitimate: answers don't change, only the next query is slower.

**No answer without a trace.**  
An untraced answer is unrepresentable — in the core and in every SDK. Nothing is fabricated.

## How it differs from a normal database

|  | A typical database | Kivi |
| --- | --- | --- |
| What is durable | Current state; UPDATE destroys the past | The events; state is recompiled from them |
| Answers | Naked values | Value + trace + scope; `why` fetches the receipts |
| Point-in-time | Snapshots and extra machinery | Free by design: fold up to record N (`--as-of`) |
| Integrity | Assumed | Verified: SHA-256 chain + Ed25519 seals; clients re-verify the server |
| Deletion | Row gone, story gone | Crypto-erase: key destroyed, bytes unchanged, erasure receipted |

## Quick start

Run the server (single static binary, or the container image):

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# or the binary
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

Append an event and read it back — the answer comes with its trace:

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # re-hash the chain, check every seal
kivi why    ./kivi.ledger 0      # the receipt behind the answer
```

Clients speak the same contract in **six languages** — Go, Python, Java & Kotlin, Rust, Node.js and .NET — with typed "entity in, entity out" access and client-side verification on by default.

## What's inside

**Traced answers**  
Every value cites the events that established it; an untraced answer cannot be constructed.

**Time travel**  
"What did we know at record N?" is one bounded replay — no snapshots, no migrations (`--as-of`).

**Verified integrity**  
SHA-256 hash chain + Ed25519 seals; a single flipped byte is caught and named.

**Crypto-erase**  
Per-record keys; erasing destroys the key, keeps the chain valid, and files the erasure itself.

**ACID, append-only form**  
Per-event atomicity, serialized writes, snapshot reads, group-commit `fsync` durability.

**Cluster & federation**  
Failover with majority commit, read replicas, and mutual witnessing between independent ledgers.

**Admin panel & identity**  
Embedded UI, receipted config, users/roles/sessions and named, revocable API keys.

**LLM door (MCP)**  
Gives an AI agent a memory that cites its sources; a missing fact is refused, not hallucinated.

**Multi-tenant, on demand**  
Several isolated tenants in one process — and a new one can be opened at runtime through the admin API, receipted, no restart.

**Runs as a service**  
systemd on Linux (.deb/.rpm) and a real Windows service from the installer — start on boot, graceful drain on stop.

## Admin panel

Every server ships an embedded web panel (no separate deployment) on the ops port. Open `https://<host>:4742/admin/` and sign in with a user account — from there, live and without touching the data plane, you can:

-   **Overview & Periods** — health, record and seal counts, rotation.
-   **Records & Search** — a live stream of appends, plus semantic search.
-   **Cluster & Federation** — topology, failover and witnessing status.
-   **Users & API Keys** — create accounts and roles, issue and revoke keys.
-   **Config** — retune limits and settings; every change is receipted into the ledger (secrets never are).

## Architecture

One process, two network doors, a strict dependency rule: the core uses the Go standard library only; gRPC lives at the edge. On disk there is exactly one kind of durable truth — the event log — and everything else is disposable.

CLI · six-language SDKs untrusting: verify client-side Admin panel · LLM agents (MCP) embedded UI · cited memory gRPC data plane · :4741 ops HTTP · :4742 core (Go stdlib): ledger · views · trace · seal · identity answers compiled at query time — nothing derived is primary disk: append-only event log (+ seals)

## Client SDKs

Six client SDKs speak the same wire contract — all MIT-licensed and published to their language’s registry. The kivi server and core stay proprietary; the clients are open.

| Language | Registry | Install | Package page |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### The untrusting client

Every SDK exposes the same small, typed surface — and client-side verification is **on by default**: the client re-hashes the chain and checks the Ed25519 seals as it reads, so a lying server or a single flipped byte is caught, not trusted.

-   `append(type, body)` — add an event, get its receipt.
-   `table(subject, attribute)` — a **traced answer**: `value` + `trace` (the record numbers that established it) + `scope`. A missing cell is an honest refusal — a native exception, never a fabricated null.
-   `table(…, as_of=N)` — the same answer **as of record N** (time travel).
-   `why(trace)` — the receipt records themselves.
-   `replay()` — stream every record, re-verified client-side.
-   `similar(query, k)` — traced semantic search (record + score).
-   `login(user, pw)` — a role-scoped session token; `with_bearer` reuses the channel under another identity.
-   `head()` — cheap orientation (head number + hash), no audit.

### Example (Python)

The shape is identical in all six languages:

```
from kivi import KiviClient

c = KiviClient("localhost:4741", token="…")     # verify=True by default
c.append("property", {"subject": "dog", "attribute": "sound", "value": "bark"})
a = c.table("dog", "sound")              # TracedAnswer(value="bark", trace=[0], scope=0)
old = c.table("dog", "sound", as_of=41)  # the answer as of record 41
receipts = c.why(a.trace)                # the actual ledger records
for rec in c.replay():                   # hash + chain + seal verified CLIENT-SIDE
    ...
```

One conformance exam runs against all six, so every language returns the same answer with the same trace. Per-language examples (Go, Java / Kotlin, Rust, Node.js, .NET) and the typed entity layer are in the [full reference](reference.en.html).

## Full documentation & source

-   **[Full reference (English)](reference.en.html)** — every feature, the CLI and gRPC API, environment variables, how-to guides, and language-tabbed code examples.
-   **Compatibility:** the gRPC API, on-disk format v2 and SDK surface are a SemVer contract — they do not break without a major-version bump.
-   **Honest status:** not yet independently security-audited; no production mileage yet. These are disclosures, not caveats on the stability promise.

## Contact

Questions, a demo, or a design-partner conversation — reach us:

-   **Email** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[Contact](#contact) · Kivi v1.1.0 · disk format v2 (frozen contract, golden vectors) · Go stdlib core, gRPC at the edges · one static binary. © iwasoft.
