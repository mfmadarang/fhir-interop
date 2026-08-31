<div align="center">

# fhir-interop

**A small clinical-data interoperability pipeline in Go.**
It reads FHIR R4 and HL7v2 into one internal model, validates it against FHIR rules, and stores it in Postgres so you can query it over GraphQL or a small REST API.

[![CI](https://github.com/mfmadarang/fhir-interop/actions/workflows/ci.yml/badge.svg)](https://github.com/mfmadarang/fhir-interop/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-learning%20project-orange)

</div>

---

## About

I built this after my internship, where I first got to use Go, GraphQL (gqlgen),
and GORM. I wanted to keep practicing that stack on something of my own instead
of letting it go stale, so I picked a domain close to my program (Health
Informatics) and built a small pipeline for parsing, validating, and querying
FHIR patient data.

Data comes in as either FHIR R4 JSON bundles or raw HL7v2 messages. Both are
converted to the same internal type as the first step, so everything after
that (validation, storage, and the two APIs) behaves the same no matter where
a record originally came from.

> **Note:** this is a learning project, not a production system. It has only
> ever seen synthetic (Synthea-generated) data. See [Notes](#notes).

## How it works

```mermaid
flowchart LR
    A[FHIR JSON Bundle] --> P["fhir.ParseBundle"]
    G[HL7v2 Message] --> H["convert package<br/>(ADT, ORU)"]
    P --> C[ParsedBundle]
    H --> C
    C --> V["Validator<br/>(internal/validate)"]
    V --> D[("Postgres<br/>via GORM")]
    D --> E["GraphQL API<br/>(gqlgen)"]
    D --> R["REST API<br/>(internal/rest)"]
    E --> F[Client / Playground]
    R --> F
```

A FHIR Bundle and an HL7v2 message both end up as the same `ParsedBundle`, so
validation, persistence, and both APIs don't care which format the data came
from.

## Features

**Parsing and conversion**
- FHIR R4 Bundle parsing for `Patient`, `Encounter`, and `Observation`
- HL7v2 to FHIR conversion (`internal/convert`): ADT^A01 (admit), ADT^A03
  (discharge), and ORU^R01 (results), mapped onto the same structs as the FHIR
  JSON path
- A hand-rolled HL7v2 parser (`internal/hl7v2`), no third-party library

**Validation**
- Structural and value checks: required fields, FHIR value sets, date formats
- Terminology validation for LOINC and SNOMED-CT codes on `Observation.code`
  against the public [tx.fhir.org](https://tx.fhir.org) server
  (`CodeSystem/$validate-code`), run through a bounded worker pool

**Storage and APIs**
- Postgres persistence via GORM, upsert by id, original resource kept in a raw
  JSONB column
- GraphQL API ([gqlgen](https://gqlgen.com/)) with a playground, static API key
  auth on `/query`, and cursor-based pagination on `patients`
- FHIR-style REST read and search for `Patient`: `GET /fhir/Patient/{id}` and
  `GET /fhir/Patient?family=...`, returning stored FHIR JSON and a `searchset`
  Bundle

**Tooling and ops**
- `cmd/ingest` for bulk-loading a folder of FHIR bundles or HL7v2 messages
- Structured logging (`log/slog`, text or JSON), `/healthz`, `/readyz`, and a
  Prometheus `/metrics` endpoint
- CI on every push and PR (gofmt, build, vet, test)
- Multi-stage Dockerfile, images published to GHCR on release

## Getting started

### Prerequisites

- Go 1.25+
- Docker (for Postgres, or bring your own instance)

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

### 2. Configure

The server reads its config from environment variables on startup
(`internal/config`) and won't start without the required ones.

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | | Postgres connection string |
| `API_KEY` | yes | | gates the GraphQL `/query` endpoint |
| `PORT` | no | `8080` | HTTP port |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | no | `text` | `text` or `json` |

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/fhir_interop?sslmode=disable"
export API_KEY="your-secret-key"
```

### 3. Load sample data

```bash
go run ./cmd/ingest                                 # every .json bundle in testdata/
go run ./cmd/ingest path/to/bundles                 # a different folder
go run ./cmd/ingest --format hl7v2 testdata/hl7v2   # HL7v2 messages instead
```

### 4. Run the server

```bash
go run ./cmd/server
```

| Path | What it is |
| --- | --- |
| `/` | GraphQL playground (open) |
| `/query` | GraphQL endpoint (needs the API key) |
| `/fhir/Patient/...` | FHIR REST read and search |
| `/app/` | small patient browser UI |
| `/demo` | pipeline visualizer |
| `/healthz`, `/readyz` | liveness and readiness |
| `/metrics` | Prometheus metrics |

## Using the APIs

### GraphQL

The playground at `http://localhost:8080/` is open, but running a query needs
the key. Add an `Authorization: Bearer <your key>` header in the playground's
**Headers** panel, or use `curl`:

```bash
curl -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ patients(first: 5) { edges { node { id familyName gender } cursor } pageInfo { hasNextPage endCursor } } }"}' \
  http://localhost:8080/query
```

```graphql
query {
  patients(first: 5) {
    edges {
      node { id familyName gender }
      cursor
    }
    pageInfo { hasNextPage endCursor }
  }
}
```

To page, pass the previous response's `endCursor` back as `after`:

```graphql
query {
  patients(first: 5, after: "<endCursor>") {
    edges { node { id familyName } }
    pageInfo { hasNextPage endCursor }
  }
}
```

### FHIR REST (Patient)

So far this only covers `Patient`, and it's read-only. These endpoints aren't
gated by the API key.

```bash
# read one patient (returns the stored FHIR Patient JSON)
curl http://localhost:8080/fhir/Patient/<id>

# search (returns a FHIR Bundle, type: searchset)
curl "http://localhost:8080/fhir/Patient?family=Smith&gender=female"
```

The search params are `family` and `given`, which match case-insensitively from
the start of the name, plus `gender` and `birthdate`, which have to match
exactly. `_count` caps how many results come back (50 by default, 200 at most).
Asking for a patient that doesn't exist returns a 404. Error responses are
plain HTTP text for now; turning them into FHIR `OperationOutcome` documents is
a later change.

### Observability

`/healthz` returns OK for as long as the process is running. `/readyz` waits
until the database answers a ping before it returns OK, and gives a 503 until
then. `/metrics` exposes the usual Prometheus counters. None of the three are
written to the request log, which keeps health checks and metric scrapers from
burying the actual traffic.

## Running with Docker

```bash
docker build -t fhir-interop .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5433/fhir_interop?sslmode=disable" \
  -e API_KEY="your-secret-key" \
  fhir-interop
```

## Testing

```bash
go test ./...
```

Tests are unit-level and don't need a database. Code that talks to Postgres
sits behind a small interface so it can be faked in tests.

## Test data

Most of the test data is synthetic FHIR bundles generated by
[Synthea](https://github.com/synthetichealth/synthea). There's also a small
hand-written fixture (`testdata/sample_minimal.json`) that keeps the unit tests
fast. Synthea doesn't export HL7v2, so the files in `testdata/hl7v2/` are
hand-written to match the structure a real ADT or ORU feed would use. No real
patient data is used anywhere in the project.

## Notes

This exists so I can practice FHIR interoperability concepts. It isn't built to
be a production system.

- Only synthetic (Synthea) data has ever touched this. Don't point it at real
  patient data.
- Not audited, not intended for clinical use.
- The GraphQL `/query` endpoint is behind a static API key. The playground
  itself isn't gated, only query execution.
- Terminology validation depends on the public `tx.fhir.org` server, which is
  explicitly non-production and rate-limited. If it's slow or unreachable,
  codes are marked unverified instead of blocking ingest (fail-open).
- HL7v2 support covers ADT^A01, ADT^A03, and ORU^R01 only.

## Maintainers

Built and maintained by [@mfmadarang](https://github.com/mfmadarang) and
[@farrahleanne](https://github.com/farrahleanne).

## License

MIT. See [LICENSE](LICENSE).
