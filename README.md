# fhir-interop

![CI](https://github.com/mfmadarang/fhir-interop/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go)

A clinical data interoperability tool. Parses and validates FHIR
patient records, persists them to Postgres, and exposes them through
a GraphQL API.

## About

I built this after wrapping up my internship, where I got hands-on experience with Go, GraphQL (gqlgen), and GORM. Wanted to
keep using that stack on something of my own instead of letting it go
stale, so I picked a domain that's actually close to my program, Health Informatics,
and built a small pipeline for parsing, validating, and querying FHIR
patient data.

## How it works

```mermaid
flowchart LR
    A[FHIR JSON Bundle] --> P["fhir.ParseBundle"]
    G[HL7v2 Message] --> H["convert package<br/>(ADT, ORU)"]
    P --> C[ParsedBundle]
    H --> C
    C --> V["Validator<br/>(internal/validate)"]
    V --> D[("Postgres<br/> via GORM")]
    D --> E["GraphQL API<br/> (gqlgen)"]
    E --> F[Client / Playground]
```

A FHIR Bundle or an HL7v2 message both end up as the same `ParsedBundle`
type, so validation, persistence, and the GraphQL API don't care which
format the data came from.

## Features

- FHIR R4 Bundle parsing for `Patient`, `Encounter`, and `Observation`
- HL7v2 → FHIR conversion (`internal/convert`): ADT^A01 (admit),
  ADT^A03 (discharge), and ORU^R01 (results) messages, mapped onto
  the same `Patient`/`Encounter`/`Observation` structs as FHIR JSON
- Structural and value validation (required fields, FHIR value sets,
  date formats)
- Terminology validation for LOINC and SNOMED-CT codes on `Observation.code`, checked against the public [tx.fhir.org](https://tx.fhir.org) terminology server (`CodeSystem/$validate-code`), run concurrently with a bounded worker pool
- Postgres persistence via GORM, with upsert semantics
- GraphQL API (built with [gqlgen](https://gqlgen.com/)) for querying
  stored records
- FHIR-style REST read and search for `Patient`: `GET /fhir/Patient/{id}`
  and `GET /fhir/Patient?family=...`, returning the stored FHIR JSON and a
  `searchset` Bundle (`internal/rest`)
- Static API key authentication on the GraphQL query endpoint (`internal/auth`)
- Cursor-based pagination on `patients` (Relay-style `Connection`/`edges`/`pageInfo`), so listing works safely as the dataset grows
- `cmd/ingest` for bulk-loading a folder of FHIR bundles
- CI pipeline (build, vet, test, gofmt check) on every push and PR
- Dockerfile + GitHub Actions workflow publishing images to GHCR on
  release
- Structured logging with `log/slog` (text or JSON), plus `/healthz`,
  `/readyz`, and a Prometheus `/metrics` endpoint (`internal/obs`)

## Getting started

### Prerequisites

- Go 1.22+
- Docker (for Postgres, or point at your own instance)

### 1. Clone and start Postgres

```bash
git clone https://github.com/mfmadarang/fhir-interop.git
cd fhir-interop

docker run --name fhir-interop-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=fhir_interop \
  -p 5433:5432 \
  -d postgres:16
```

### 2. Set the database URL

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/fhir_interop?sslmode=disable"
```

### 3, Set the API key

```bash
export API_KEY="your-secret-key"
```

The server reads its config from env vars on startup (`internal/config`).
`API_KEY` and `DATABASE_URL` are required and it won't start without
them. Optional ones:

| Var | Default | Notes |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | `text` | `text` or `json` |

### 4. Load sample data

```bash
go run ./cmd/ingest
```

Loads every `.json` bundle in `testdata/` by default, or pass a
different directory: `go run ./cmd/ingest path/to/bundles`.

To load HL7v2 messages instead of FHIR JSON:

```bash
go run ./cmd/ingest --format hl7v2 testdata/hl7v2
```

### 5. Run the server

```bash
go run ./cmd/server
```

Opens a GraphQL playground at `http://localhost:8080/`. The playground UI itself is open, but queries against `/query` require the key, add an `Authorization: Bearer <your key>` header in the playground's **Headers** panel, or with `curl`:

The server also exposes `/healthz` (always OK while the process is up), `/readyz` (OK once the database answers a ping, 503 otherwise), and `/metrics` (Prometheus format). These three aren't logged, so health and scrape traffic doesn't flood the output.

```bash
curl -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ patients(first: 5) { edges { node { id familyName } cursor } pageInfo { hasNextPage endCursor } } }"}' \
  http://localhost:8080/query
```

```graphql
query {
  patients(first: 5) {
    edges {
      node {
        id
        familyName
        gender
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

To fetch the next page, pass the previous response's `endCursor` as `after`:

```graphql
query {
  patients(first: 5, after: "<endCursor from previous page>") {
    edges {
      node {
        id
        familyName
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

### FHIR REST API (Patient)

A small FHIR-style REST API sits next to the GraphQL one. It's read-only and
currently only covers `Patient`. No API key needed.

Read one patient by id (returns the stored FHIR `Patient` JSON):

```bash
curl http://localhost:8080/fhir/Patient/<id>
```

Search (returns a FHIR `Bundle` with `type: searchset`):

```bash
curl "http://localhost:8080/fhir/Patient?family=Smith&gender=female"
```

Supported search params: `family`, `given` (both case-insensitive prefix
match), `gender`, `birthdate` (both exact), and `_count` to limit the number
of results (default 50, max 200). A missing patient returns 404. Errors are
plain HTTP for now; a later change adds FHIR `OperationOutcome` responses.

### Running via Docker

```bash
docker build -t fhir-interop .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5433/fhir_interop?sslmode=disable" \
  fhir-interop
```

## Testing

```bash
go test ./...
```

## Test data

Uses [Synthea](https://github.com/synthetichealth/synthea)-generated
synthetic FHIR bundles for development and testing, plus a small
hand-crafted fixture (`testdata/sample_minimal.json`) for fast unit
tests. Synthea doesn't export HL7v2, so `testdata/hl7v2/` is hand-
written instead, following the same message structure a real ADT/ORU
feed would use. No real patient data is used anywhere in this project.

## Notes

This is a learning project built to practice FHIR interoperability
concepts, not a production-ready system. A few things worth knowing
before using it as more than a reference:

- Only synthetic (Synthea-generated) data has ever touched this
  project. Don't point it at real patient data.
- The GraphQL endpoint (`/query`) is gated behind a static API key (`API_KEY` env var). The playground itself is not gated, only query execution.
- Not audited, not intended for clinical use.
- Terminology validation depends on the public `tx.fhir.org` server, which is explicitly non-production and rate-limited; if it's unreachable or slow, codes are marked unverified rather than blocking ingest (fail-open behavior).
- HL7v2 support covers ADT^A01, ADT^A03, and ORU^R01 only.