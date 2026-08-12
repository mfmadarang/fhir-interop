# fhir-interop

A clinical data interoperability tool. Parses and validates FHIR patient
records, persists them, and exposes them through a GraphQL API.

## Scope (v1)

- Parse FHIR JSON resources (Patient, Encounter, Observation)
- Validate structure and required fields
- Persist validated records to Postgres via GORM
- Expose records through a GraphQL API (gqlgen)

## Roadmap

- v1: FHIR-only, as scoped above
- v2 (stretch): HL7v2 message parsing and HL7v2 -> FHIR conversion

## Test data

Uses [Synthea](https://github.com/synthetichealth/synthea)-generated
synthetic FHIR bundles for development and testing. No real patient data
is used anywhere in this project.

## Layout

```
cmd/server/       entrypoint
internal/fhir/     FHIR resource structs + parsing
internal/validate/ validation logic
internal/store/    Postgres + GORM persistence
internal/graph/    GraphQL schema + resolvers (gqlgen)
testdata/          sample Synthea FHIR bundles
```
