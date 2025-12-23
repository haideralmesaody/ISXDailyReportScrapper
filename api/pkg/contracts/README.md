# Contracts Package

This package contains the Single Source of Truth (SSOT) for all data structures, API contracts, and domain models used throughout the ISX Daily Reports Scrapper application.

## Directory Structure

- **domain/** - Core business domain models (reports, tickers, companies, etc.)
- **api/v1/** - Versioned API request/response contracts
- **events/** - WebSocket and event message contracts
- **database/** - Database schemas and migrations
- **generated/** - Auto-generated code from various tools

## Principles

1. **Single Source of Truth**: All shared types must be defined here
2. **Code Generation**: Use tools like sqlc, oapi-codegen, and go2ts
3. **Versioning**: API contracts are versioned (v1, v2, etc.)
4. **No Business Logic**: Only data structures and interfaces
5. **Clear Documentation**: Every struct and field must be documented

## Usage

Import contracts in your service code:

```go
import (
    "isxcli/pkg/contracts/domain"
    apiv1 "isxcli/pkg/contracts/api/v1"
    "isxcli/pkg/contracts/events"
)
```

## Code Generation

Run code generation after modifying contracts:

```bash
task generate
```

## Change Log
- 2025-07-31: Renamed pipeline to operations for business-friendly terminology
- 2025-07-31: Updated all references from stages to steps
- 2025-07-31: Simplified implementation - removed unnecessary adapter pattern and duplicate types
- 2025-07-30: Added operations domain contracts for Phase 4 implementation
  - Added domain/operations.go with Operation, OperationStep, and OperationConfig types
  - Added api/v1/operations.go with request/response contracts
  - Updated events/websocket.go with operations-specific message types
- 2025-01-30: Renamed WebSocket events from operation/step to operations/step terminology (Task 2.3)
  - Updated event message types: pipeline_progress → operation_progress, etc.
  - Updated channel types: ChannelTypePipeline → ChannelTypeOperations
  - Added backward compatibility aliases for existing integrations
  - Updated TypeScript types to match new terminology
- 2025-01-29: Removed machine ID binding from license contracts (simplified license system)