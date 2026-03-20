## Why

The Go rewrite already reaches initial registration signaling, but it does not yet define a stable, end-to-end happy path for full UE registration and PDU session establishment. Without a narrow target, the rewrite risks drifting into piecemeal protocol work, unnecessary parity chasing, and unreadable ports of C++ behavior.

## What Changes

- Define a constrained happy-path flow for the Go implementation covering UE bootstrap, registration, security setup, registration completion, PDU session establishment, and basic user-plane readiness.
- Document which parts of the flow should remain handwritten state-machine logic versus generated protocol structures and message helpers.
- Identify which C++ behaviors and protocol edge cases are explicitly out of scope for the Go rewrite at this stage.
- Establish a verification path against the existing live core network functions running on the machine, using free5GC or equivalent Docker-hosted NFs for the supported happy path.
- Clarify the implementation boundary so future work can extend the Go rewrite incrementally without reopening full C++ feature parity.

## Capabilities

### New Capabilities
- `go-happy-path-session-flow`: Defines the supported Go rewrite behavior for registration and PDU session establishment, including supported message flow, implementation boundaries, and explicit non-goals.

### Modified Capabilities

## Impact

- Affected code: `go/internal/ue`, `go/internal/gnb`, `go/internal/nas`, `go/internal/ngap`, `go/internal/rrc`, `go/internal/gtp`, and related config and verification docs.
- Affected systems: Go UE and gNB binaries, free5GC-based registration verification, and future code generation decisions for protocol helpers.
- Scope impact: deprioritizes full C++ parity, advanced RRC features, and historical edge-case behavior in favor of a readable, narrow vertical slice.
