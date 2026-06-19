# ShortLink

A URL shortening service written in Go. It exposes two JSON endpoints, `/encode`
and `/decode`, and persists mappings to an embedded key/value store so that
previously encoded URLs can still be decoded after a restart.

## Design

The code is organised in clear layers, each behind an interface so they can be
swapped or tested in isolation:

```
cmd/server            process entrypoint, configuration, graceful shutdown
internal/handler      HTTP transport: routing, middleware, request/response
internal/service      business logic: validation, encode/decode, code policy
internal/store        persistence interface + bolt and in-memory implementations
internal/shortcode    cryptographically random base62 code generator
internal/config       environment based configuration
```

The HTTP layer depends on a small `Encoder` interface, and the service depends on
a `Store` interface. This keeps the layers decoupled and makes the storage engine
an implementation detail.

### How encoding works

1. The URL is validated and normalised (absolute, `http`/`https`, non-empty host).
2. The service asks the store to persist the mapping, passing a code generator.
3. The store runs the whole operation in a single transaction:
   - if the URL was already shortened, it returns the existing code (idempotent);
   - otherwise it generates a random base62 code, checks it is free, and writes
     both the `code -> url` and `url -> code` mappings atomically.

Because the read, the collision check, and the write happen inside one
transaction, concurrent requests for the same URL always converge on a single
code instead of racing and producing duplicates.

### Short codes

Codes are 6 characters from a 62 character alphabet (`0-9A-Za-z`), which gives
about 56.8 billion combinations. They are produced with `crypto/rand` using
rejection sampling to avoid modulo bias. If a generated code collides, the store
retries; after a few attempts it generates a longer code.

## Requirements

- Go 1.24 or newer
- Docker (optional, for containerised runs)

## Running locally

```bash
go run ./cmd/server
```

The server listens on `http://localhost:8080` by default and writes its database
to `shortlink.db` in the working directory.

Configuration is via environment variables:

| Variable             | Default                 | Description                          |
| -------------------- | ----------------------- | ------------------------------------ |
| `PORT`               | `8080`                  | Listen port                          |
| `BASE_URL`           | `http://localhost:8080` | Prefix used when building short URLs  |
| `DB_PATH`            | `shortlink.db`          | Path to the database file            |
| `CODE_LENGTH`        | `6`                     | Length of generated short codes      |
| `RATE_LIMIT_PER_MIN` | `120`                   | Per-IP request budget per minute     |

Example:

```bash
PORT=9090 BASE_URL="http://localhost:9090" go run ./cmd/server
```

## Running the tests

```bash
go test ./... -race -cover
```

The suite covers the short code generator, the service logic, the HTTP endpoints,
and the storage layer, including a test that proves mappings survive a restart and
a concurrency test (under `-race`) that proves encoding the same URL from many
goroutines yields one consistent code.

## API

### `POST /encode`

Request:

```json
{ "url": "https://codesubmit.io/library/react" }
```

Response `200 OK`:

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

Errors: `400` invalid URL or malformed body, `413` body too large.

```bash
curl -s -X POST http://localhost:8080/encode \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://codesubmit.io/library/react"}'
```

### `POST /decode`

Request:

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

Response `200 OK`:

```json
{ "url": "https://codesubmit.io/library/react" }
```

Errors: `400` invalid input, `404` unknown code.

```bash
curl -s -X POST http://localhost:8080/decode \
  -H 'Content-Type: application/json' \
  -d '{"short_url":"http://localhost:8080/GeAi9K"}'
```

A bare code (for example `GeAi9K`) is also accepted in the `short_url` field.

### `GET /healthz`

Liveness probe, returns `{"status":"ok"}`.

## Running with Docker

```bash
docker build -t shortlink .
docker run --rm -p 8080:8080 -v shortlink-data:/data shortlink
```

The database lives on the `/data` volume so it survives container restarts.

## Deployment

The image is a small static binary on a distroless base running as a non-root
user, so it can be deployed to any container host (Render, Fly.io, Railway,
Cloud Run, etc.). Point `DB_PATH` at a mounted volume and set `BASE_URL` to the
public hostname.

## Security

This service only stores and returns strings as JSON. It never fetches the target
URL and never issues an HTTP redirect, which keeps the attack surface small.

Considered and addressed:

- **Oversized payload / memory exhaustion** — every request body is capped with
  `http.MaxBytesReader`, and the JSON decoder rejects unknown fields and trailing
  data.
- **Invalid or non-HTTP input** — URLs are parsed and must be absolute with an
  `http`/`https` scheme and a host; everything else is rejected with `400`.
- **Code enumeration** — codes are generated with `crypto/rand`, not a sequential
  counter, so the keyspace cannot be walked to scrape stored URLs.
- **Lookup injection** — decode only accepts `[0-9A-Za-z]`, so unexpected input
  is rejected before it reaches storage.
- **Request floods** — a per-IP token bucket rate limiter caps request volume.
- **Panics leaking internals** — recovery middleware turns panics into generic
  `500` responses; error bodies never expose internal details.

Out of scope for the current design, but documented for completeness: if a
redirecting endpoint (`GET /{code}` returning a 3xx) is added later, then SSRF and
open-redirect become real concerns and would require blocking private/loopback
address ranges and rejecting self-referential targets.

## Scalability

The current implementation uses a single embedded database, which is intentional:
it is simple, dependency-free, and survives restarts, which is all the assignment
requires. It is bound to a single process. The following is how it would grow.

### Collisions

With 6-character codes there are ~56.8 billion possibilities. The probability of a
collision stays negligible until the store holds a very large number of URLs, and
when codes do collide the store simply retries and, if needed, lengthens the code.
There are two strategies worth comparing:

- **Random + collision check (used here).** Codes are unguessable, but each write
  must verify the code is free. At very high scale the check cost grows and a
  longer code length is used to keep the space sparse.
- **Monotonic counter / Snowflake ID encoded as base62.** Collisions become
  impossible by construction and writes need no check, but codes become
  sequential and therefore enumerable, and a distributed counter needs
  coordination (for example a Snowflake-style ID with per-node bits).

The keyspace is unguessable here by design, so the random approach was chosen and
the counter approach is the documented alternative for a coordination-heavy,
write-heavy deployment.

### Storage and throughput

- The `Store` interface is the seam for scaling: swapping the embedded database
  for PostgreSQL or a distributed KV store (DynamoDB, Cassandra) requires no
  changes to the service or HTTP layers.
- The workload is read-heavy (decodes greatly outnumber encodes), so a
  read-through cache (Redis/Memcached) in front of decode would absorb most
  traffic.
- The service is stateless apart from storage, so it can be replicated behind a
  load balancer and scaled horizontally; shard or partition the data store by
  code prefix or hash when a single primary is no longer enough.
