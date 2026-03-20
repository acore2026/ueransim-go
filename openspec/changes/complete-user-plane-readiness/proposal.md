## Why

The Go rewrite already completes happy-path registration and reaches `PDU Session Establishment Accept`, but the session is not yet usable for traffic. The remaining gap is not generic protocol breadth; it is the specific control-plane and data-plane work required to turn that accepted session into a basic working user plane against the live NF environment.

## What Changes

- Extend the documented Go happy path from "session accepted" to "session usable for basic traffic".
- Parse the PDU session resource content carried in `InitialContextSetupRequest` rather than extracting only the NAS PDU.
- Create real gNB-side PDU session state for UE association, tunnel parameters, and QoS-flow-to-session mapping needed by the supported slice.
- Replace the placeholder gNB GTP task with a minimal real GTP-U path for the supported happy path.
- Wire UE TUN traffic into the established session and deliver downlink GTP traffic back to the UE.
- Send the required NGAP bearer/setup response so the RAN side of the happy-path setup is complete.
- Update verification guidance so success requires observable user-plane readiness, not only NAS/session-accept milestones.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `go-happy-path-session-flow`: tighten the happy-path contract so it requires basic user-plane readiness after PDU session establishment, including tunnel setup, bearer response handling, and a documented end-to-end verification path.

## Impact

- Affected code: `go/internal/gnb`, `go/internal/ue`, `go/internal/ngap`, `go/internal/gtp`, and related verification docs.
- Affected behavior: the Go happy path now covers bearer/tunnel completion and basic packet flow instead of stopping at `PDU Session Establishment Accept`.
- External systems: validation continues to rely on the existing Docker-hosted free5GC NFs already running on the machine.
