# ADR 0004: Immutable API Error Catalog

- **Status**: Accepted
- **Date**: 2026-07-23

## Context & Problem
Inconsistent error responses and ad-hoc error strings across services complicate API debugging for client applications and frontend integration.

## Decision
All domain exceptions MUST return `*exception.APIError` instances created via standard codes (`NOT_FOUND`, `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `DATABASE_ERROR`, `INTERNAL_ERROR`). Immutable error sentinels are cataloged in `internal/foundation/exception/sentinels.go` and implement Go 1.13 `errors.Is` comparison.

## Consequences
- Every error returned to API consumers follows a single, predictable JSON format: `{"error": {"code": "...", "message": "..."}}`.
- Error status codes are automatically mapped to standard HTTP response status codes (400, 401, 403, 404, 500).
