<!--
Sync Impact Report:
- Version change: 0.0.0 -> 1.0.0 (Initial Draft)
- List of modified principles:
  - [PRINCIPLE_1_NAME] -> I. Clean Code & Idiomatic Go
  - [PRINCIPLE_2_NAME] -> II. Test-Driven Development (TDD)
  - [PRINCIPLE_3_NAME] -> III. MVP Focus & Simplicity
  - [PRINCIPLE_4_NAME] -> IV. Protocol Integrity (3GPP Compliance)
  - [PRINCIPLE_5_NAME] -> V. Task-Based Actor Concurrency
- Added sections:
  - Technical Constraints
  - Development Workflow
- Templates requiring updates:
  - .specify/templates/plan-template.md (✅ updated)
  - .specify/templates/spec-template.md (✅ updated)
  - .specify/templates/tasks-template.md (✅ updated)
-->

# UERANSIM-Go Constitution

## Core Principles

### I. Clean Code & Idiomatic Go
Code MUST be idiomatic, following standard Go conventions (gofmt, go vet). Focus on readability and maintainability. Avoid complex abstractions where simple ones suffice. Use structured logging and clear error handling. Rationale: Maintainability is key for a long-term rewrite project.

### II. Test-Driven Development (TDD)
Every new feature or bug fix MUST be accompanied by tests. Write failing tests first to define behavior, then implement the minimal code to make them pass. Aim for high unit test coverage of protocol codecs and core logic. Rationale: Ensures correctness and prevents regressions in complex protocol implementations.

### III. MVP Focus & Simplicity
Prioritize functionality that delivers a Minimum Viable Product. Implement only what is necessary for the current task. Avoid "just-in-case" features or over-engineered architectures. YAGNI (You Ain't Gonna Need It) is the default stance. Rationale: Prevents scope creep and keeps the project focused on core value.

### IV. Protocol Integrity (3GPP Compliance)
Strictly adhere to 3GPP specifications for NGAP, NAS, and RRC. Ensure bit-level compatibility with existing 5G Core implementations (Open5GS, free5GC). Codecs must be robust and well-tested against spec-defined edge cases. Rationale: Interoperability is the primary goal of the simulator.

### V. Task-Based Actor Concurrency
Maintain the established task-based actor model for concurrency. Each protocol layer or service should be a self-contained task communicating via asynchronous messages. This ensures isolation and simplifies debugging of complex state machines. Rationale: Provides a structured way to handle the asynchronous nature of telecommunications protocols.

## Technical Constraints

- **Language**: Go 1.26.0+
- **Dependencies**: Minimize external dependencies; prefer standard library or well-vetted protocol libraries (e.g., free5gc).
- **Performance**: High-throughput User Plane (GTP-U/TUN) must be optimized for low latency and high bandwidth.

## Development Workflow

- **Feature Lifecycle**: Define specification -> Implement via TDD -> Verify against integration tests.
- **Continuous Refactoring**: Refactoring is a mandatory part of the Red-Green-Refactor cycle to prevent technical debt.
- **Documentation**: Code should be self-documenting; complex protocol logic must be accompanied by references to relevant 3GPP TS sections.

## Governance

This constitution supersedes all other development practices within the `go/` subtree. Amendments require a version bump and updated documentation. All pull requests and code changes must be validated against these principles. Use `go/README.md` for specific implementation details.

**Version**: 1.0.0 | **Ratified**: 2026-03-14 | **Last Amended**: 2026-03-14
