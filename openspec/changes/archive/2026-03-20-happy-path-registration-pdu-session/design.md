## Context

The Go rewrite already provides transport, codec foundations, and a partial registration path, but it does not yet define a narrow end-to-end target for readable implementation. The current risk is that work will continue as isolated protocol fixes or ad hoc ports from the C++ codebase, which would slow the rewrite and import complexity that the project no longer wants to preserve.

The change is intentionally scoped around coexistence. The C++ implementation remains available for broad protocol coverage and historical edge cases, while the Go implementation focuses on a stable vertical slice that is easy to reason about, test, and extend.

## Goals / Non-Goals

**Goals:**
- Establish one supported Go happy path from UE startup through registration and PDU session establishment.
- Make the implementation boundary explicit so engineers know which logic is handwritten and which protocol artifacts are generated.
- Prefer readable state-machine code over feature parity with the C++ tree.
- Declare legacy edge cases, advanced RRC depth, and feature-complete parity as out of scope for this slice.
- Define a concrete verification target against free5GC.

**Non-Goals:**
- Full parity with the existing C++ implementation.
- Support for handover, recovery, abnormal retries, multi-session orchestration, or advanced radio behavior.
- Reproducing every ASN.1 helper or protocol convenience from the C++ tree.
- Solving runtime-model concerns beyond what is required to deliver the happy path.

## Decisions

### Decision: Build one narrow vertical slice before expanding protocol breadth

The Go rewrite will target a single happy path:
UE bootstrap -> RRC setup -> Registration Request -> Authentication -> Security Mode -> Registration Accept/Complete -> PDU Session Establishment -> user-plane readiness.

This prioritizes a complete slice over partial support for many procedures.

Alternatives considered:
- Port more protocol messages first. Rejected because it increases surface area without proving the full path works.
- Chase C++ parity module by module. Rejected because coexistence removes the need for parity as a near-term goal.

### Decision: Keep state transitions and control flow handwritten

UE and gNB procedure orchestration will remain explicit, handwritten logic in Go state-machine code. This includes:
- procedure state progression
- validation of expected next messages
- retries/timeouts that are part of the supported happy path
- mapping between node state and protocol exchanges

These areas are the most valuable to keep readable, debuggable, and intentionally scoped.

Alternatives considered:
- Generate full procedure logic from message schemas. Rejected because schemas do not capture the desired behavioral simplifications or supported subset cleanly.
- Port procedural control flow from C++ with minimal translation. Rejected because it would import historical complexity and reduce readability.

### Decision: Generate protocol structure helpers, not product behavior

Generated code is appropriate for protocol structures that are repetitive and schema-driven, including:
- ASN.1 or IE container bindings
- message field encoders/decoders
- typed builder helpers for stable message shapes used in the happy path

Generated code must stay behind small handwritten wrapper APIs so application code reads in terms of procedures rather than wire-format details.

Alternatives considered:
- Handwrite all message builders. Rejected because it is slow and error-prone for repetitive structures.
- Expose generated types directly throughout the codebase. Rejected because it leaks schema complexity into state-machine logic.

### Decision: Explicitly abandon non-happy-path C++ complexity for this change

The implementation will not preserve C++ behaviors that are not required by the supported slice, including:
- advanced RRC coverage beyond what the happy path needs
- feature-complete edge-case handling
- legacy recovery branches and rare failure handling
- parity-only helper layers with no clear readability benefit

This is an explicit product decision, not deferred accidental work.

Alternatives considered:
- Keep latent hooks for most C++ branches. Rejected because it creates maintenance burden before the supported slice is stable.

### Decision: Verify with the existing live NF environment on the machine

Success for this change is tied to a documented validation flow against the core network functions already running on the machine, using the existing Go binaries and Docker-hosted NFs as the primary acceptance target. Mocked cores are useful for isolated debugging, but they are not the main acceptance path for this change.

Alternatives considered:
- Delay integration verification until broader feature parity exists. Rejected because it weakens the definition of done.
- Rely primarily on mocks for registration and session establishment. Rejected because the current environment already provides real NFs and the supported slice must prove itself against them.

## Risks / Trade-offs

- [Happy-path scope may be too narrow for some users] -> Mitigation: keep the C++ implementation available and state the Go non-goals explicitly.
- [Generated helpers may still leak unreadable schema details] -> Mitigation: require handwritten wrapper APIs at the protocol boundary.
- [Dropping C++ branches may hide future extension needs] -> Mitigation: record abandoned areas explicitly so later work can reintroduce them deliberately instead of by accident.
- [A single-core integration target may mask interoperability differences] -> Mitigation: treat the currently running NF environment as the acceptance baseline for this change, not the final interoperability claim.

## Migration Plan

1. Define the supported happy-path contract in specs.
2. Implement the missing registration, security, and PDU session steps in the Go path only.
3. Add or update verification guidance for the supported flow using the live NF containers already running on the machine.
4. Leave the C++ implementation unchanged as the fallback for broader behavior.
5. Expand beyond the happy path only through follow-on changes.

Rollback is low-risk because the change is additive to the Go rewrite and does not remove the existing C++ implementation.

## Open Questions

- Which generated source of truth should back any new helper generation: existing ASN artifacts, a narrower local schema, or wrappers over third-party library types?
- How much timeout and retry behavior belongs in the supported happy path versus follow-on reliability work?
- Should user-plane validation stop at session establishment, or must it include a basic packet exchange through TUN and GTP-U?
