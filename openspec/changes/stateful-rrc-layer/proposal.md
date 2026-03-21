## Why

The current RRC implementation in Go is a stateless relay that only supports a "Happy Path" vertical slice. It lacks the ability to manage connection lifecycles, handle failures, or transition between Idle, Connecting, and Connected states as required by the 3GPP TS 38.331 specification. Implementing a formal RRC state machine is essential for architectural depth, stability, and supporting advanced features like paging and connection re-establishment.

## What Changes

- **RRC State Machine**: Implement a formal state machine in the RRC task with `StateIdle`, `StateConnecting`, and `StateConnected`.
- **Connection Lifecycle**: Add handling for `RRCSetup`, `RRCReconfiguration`, and `RRCRelease` messages to drive state transitions.
- **NAS Integration**: Update the interaction between NAS and RRC to respect the current RRC state (e.g., buffering NAS messages while connecting).
- **Failure Handling**: Implement timers and retry logic for RRC connection establishment (T300 timer).

## Capabilities

### New Capabilities
- `rrc-state-management`: Formal management of RRC states (Idle, Connecting, Connected) and their transitions.
- `rrc-procedure-control`: State-driven control of RRC procedures including setup, reconfiguration, and release.

### Modified Capabilities
- `rrc-interop-packing`: Extend requirements to include bit-packing for `RRCReconfigurationComplete` and other lifecycle messages.

## Impact

- `go/internal/ue/rrc.go`: Transition `RrcTaskHandler` from stateless to stateful logic.
- `go/internal/rrc/builder.go`: Add bit-packers for `RRCReconfigurationComplete` and potentially others.
- `go/internal/ue/nas.go`: Update how NAS interacts with the RRC layer during connection setup.
