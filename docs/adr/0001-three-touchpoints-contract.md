# ADR 0001: The 3-Touchpoints Architectural Contract

- **Status**: Accepted
- **Date**: 2026-07-23

## Context & Problem
In fast-growing teams, developers spent up to 80% of their time writing repetitive HTTP routing, DTO request/response mappings, validator boilerplate, error response wrappers, and dependency injection wiring. Different developers often wrote module code inconsistently, leading to architectural fragmentation.

## Decision
We establish a strict framework contract: For any new module or domain feature, developers MUST only touch **3 primary areas**:
1. **Entity / Model**: Defining domain data attributes and database column specifications.
2. **Repository Query**: Writing domain-specific database queries.
3. **Business Rule**: Implementing domain usecase logic in the Service layer.

All infrastructure concerns (Fiber HTTP routing, DTO request/response schemas, DTO-entity mappers, validator parsing, error translation, logging, dependency injection wiring, and Swagger documentation) MUST be handled automatically by **Foundation** utilities and the **Code Generator**.

## Consequences
- **Positive**: 80% reduction in repetitive boilerplate; 100% consistent API contract across all domain modules.
- **Negative**: Developers must adhere to starter kit conventions rather than inventing ad-hoc folder structures.
