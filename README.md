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
    A[FHIR JSON Bundle] --> B["Parser<br/>(internal/fhir)"]
    B --> C["Validator<br/> (internal/validate)"]
    C --> D[("Postgres<br/> via GORM")]
    D --> E["GraphQL API<br/> (gqlgen)"]
    E --> F[Client / Playground]
```

A FHIR Bundle (one patient's full synthetic record) gets parsed into
typed Go structs, checked against FHIR's actual value sets and
required fields, saved to Postgres, and made queryable through
GraphQL.

## Features

- FHIR R4 Bundle parsing for `Patient`, `Encounter`, and `Observation`
- Structural and value validation (required fields, FHIR value sets,
  date formats)
- Postgres persistence via GORM, with upsert semantics
- GraphQL API (built with [gqlgen](https://gqlgen.com/)) for querying
  stored records
- `cmd/ingest` for bulk-loading a folder of FHIR bundles
- CI pipeline (build, vet, test, gofmt check) on every push and PR
- Dockerfile + GitHub Actions workflow publishing images to GHCR on
  release

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

### 3. Load sample data

```bash
go run ./cmd/ingest
```

Loads every `.json` bundle in `testdata/` by default, or pass a
different directory: `go run ./cmd/ingest path/to/bundles`.

### 4. Run the server

```bash
go run ./cmd/server
```

Opens a GraphQL playground at `http://localhost:8080/`. Try:

```graphql
query {
  patients(limit: 5) {
    id
    familyName
    gender
  }
}
```

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
tests. No real patient data is used anywhere in this project.

## Notes

This is a learning project built to practice FHIR interoperability
concepts, not a production-ready system. A few things worth knowing
before using it as more than a reference:

- Only synthetic (Synthea-generated) data has ever touched this
  project. Don't point it at real patient data.
- The GraphQL API has no authentication. Don't expose it publicly or
  connect it to anything sensitive.
- Not audited, not intended for clinical use.