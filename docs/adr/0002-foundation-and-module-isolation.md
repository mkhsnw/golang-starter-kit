# ADR 0002: Foundation Centrality and Module Isolation

- **Status**: Accepted
- **Date**: 2026-07-23

## Context & Problem
Monolithic backend codebases often fail over time because modules become tightly coupled, importing each other's data structures or leaking HTTP framework contexts into deep domain business logic.

## Decision
1. **Foundation Centrality**: Core utilities (`appcontext`, `database`, `exception`, `health`, `logger`, `mapper`, `response`, `validator`) reside in `internal/foundation/` and serve as shared building blocks.
2. **Module Isolation**: Each domain module under `internal/module/<domain>` is isolated.
3. **AST Architecture Linter Enforcement**: The system uses `gokit lint` (`cmd/lint`) to automatically reject any code where Controllers access Repositories directly or Services import web frameworks (`fiber/v3`).

## Consequences
- Clean separation of concerns with zero framework leakage in domain service layers.
- Modules can easily be extracted into independent microservices if required in the future.
