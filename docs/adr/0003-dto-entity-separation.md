# ADR 0003: Strict DTO and Domain Entity Separation

- **Status**: Accepted
- **Date**: 2026-07-23

## Context & Problem
Using database models/entities directly as HTTP request/response payloads risks exposing internal database column names, soft-delete metadata, or sensitive user fields (e.g. password hashes) over the network.

## Decision
Domain Entities (GORM models in `entity.go`) are strictly isolated to the Database/Repository layer. HTTP payloads MUST use dedicated DTO structs defined under `internal/module/<domain>/dto`. Explicit mapper functions (`mapper.go`) convert Entities into Response DTOs.

## Consequences
- Prevents accidental security leaks of sensitive database fields.
- Allows database schema evolution without breaking HTTP API contracts.
