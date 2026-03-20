## Context

The current Go rewrite keeps NAS procedure progression explicit in `go/internal/ue/nas.go`, but the message layer in `go/internal/nas` is still largely manual. That manual layer currently works for the happy path, but it couples future feature work to byte-level parsing and bespoke IE handling.

The repo already has a forked NAS library available locally at `/root/proj/go/acore-forks/nas` with module path `github.com/acore2026/nas`. That library has much broader NAS message and IE coverage than the local handwritten codec code. The design goal is to use that broader coverage without making the procedure layer harder to read.

## Goals / Non-Goals

**Goals:**
- Replace manual NAS message structure handling for the supported happy-path messages with a local adapter backed by `github.com/acore2026/nas`.
- Keep UE and gNB control-flow code readable by preserving small local helper APIs such as `DecodeAuthenticationRequest(...)` and `BuildRegistrationRequest(...)`.
- Limit the first migration to the messages required by the current verified slice: registration, authentication, identity, security mode, registration complete, UL NAS transport, DL NAS transport, and initial PDU session establishment signaling.
- Preserve current test coverage and happy-path behavior against live free5GC containers.

**Non-Goals:**
- Rewriting UE or gNB procedure logic into `free5gc/nas` types.
- Adopting the entire free5GC architecture or message model across the codebase.
- Expanding protocol scope beyond the current happy-path registration and initial session-establishment slice.
- Solving user-plane readiness as part of this change.

## Decisions

### Decision: Keep a local adapter boundary
The Go rewrite SHALL keep a local `go/internal/nas` adapter layer and SHALL NOT expose `nasMessage` and `nasType` objects directly to the procedure layer.

Rationale:
- `go/internal/ue/nas.go` is currently readable because it talks in procedure terms, not generated-type terms.
- The adapter boundary allows the project to reuse the fork's codec coverage while keeping the upper layer stable and compact.

Alternative considered:
- Directly using `github.com/acore2026/nas` in `go/internal/ue/nas.go`.
  Rejected because it would spread generated-type manipulation into the procedure logic and make further flow changes harder to read.

### Decision: Migrate only the supported happy-path NAS messages first
The first implementation pass SHALL only cover the NAS messages and IEs already used by the current registration and initial PDU-session path.

Rationale:
- The repo has already verified a narrow live flow. This change should improve maintainability of that path before expanding scope.
- A narrow migration reduces the chance of breaking the currently working control-plane slice.

Alternative considered:
- Replace the entire local NAS package at once.
  Rejected because it would increase migration risk and blur whether failures come from adapter design or protocol-scope expansion.

### Decision: Preserve local DTOs for decoded results
Decoded data consumed by the procedure layer SHALL continue to be represented as small local result structs or equivalent stable helper return types.

Rationale:
- The procedure layer only needs a small subset of each message.
- Stable local DTOs reduce lock-in to a specific external package shape and make tests simpler.

Alternative considered:
- Return `free5gc/nas` message structs directly from helper functions.
  Rejected because it couples the rest of the codebase to that package and makes future replacement or targeted wrapping harder.

### Decision: Wire the fork locally during migration
The implementation SHOULD use the forked NAS module already present on disk during migration and verification.

Rationale:
- The fork exists specifically so it can be adjusted when the current happy path exposes gaps.
- Local availability reduces integration friction while the adapter layer is being stabilized.

Alternative considered:
- Depend on upstream `free5gc/nas` directly.
  Rejected for the initial migration because the local fork is already available and intended for modification if required.

## Risks / Trade-offs

- [Adapter too thin to hide external type complexity] → Keep exported/local helper signatures small and convert external structs immediately at the adapter boundary.
- [Behavior changes during migration break the current happy path] → Migrate message-by-message, preserve existing tests, and rerun live registration validation after the adapter-backed path is in place.
- [The fork does not exactly match current wire-format expectations] → Patch the fork only when necessary and keep local tests that lock in the expected bytes for the supported flow.
- [Two NAS representations coexist for too long] → Explicitly remove migrated manual codecs once their adapter-backed replacements are verified.
- [Future contributors bypass the adapter and import the fork directly into procedure code] → Document the boundary in code comments and design, and keep package-level helpers as the intended entry points.
