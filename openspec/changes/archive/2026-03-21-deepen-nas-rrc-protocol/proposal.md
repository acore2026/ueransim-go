## Why

The current Go implementation uses a naive 8-bit NAS sequence number and a non-compliant custom RRC container, which prevents long-running sessions and interoperability with real 5G Core/gNB. Deepening these layers with 32-bit COUNT management and UPER-compliant bit-packing is essential for protocol robustness and network interoperability.

## What Changes

- **NAS Security**: Implement 3GPP-compliant 32-bit COUNT management with 24-bit overflow tracking and 8-bit rollover handling.
- **NAS Security**: Add replay protection with a 16-message sliding window for received sequence numbers.
- **RRC Layer**: Implement authentic 3GPP TS 38.331 UPER bit-packing for UL-DCCH messages, starting with `RRCSetupComplete`.
- **RRC Layer**: Remove the non-standard `SimpleRRC` TLV wrapper in favor of the authentic bit-packed implementation.
- **Verification**: Ensure the Go UE can still complete the registration and PDU session flow against live free5GC containers as described in `go/VERIFY_REGISTRATION.md`.

## Capabilities

### New Capabilities
- `nas-robust-security`: Stateful 32-bit COUNT management and replay protection for NAS protocol security.
- `rrc-interop-packing`: 3GPP TS 38.331 compliant UPER bit-packing for RRC control plane messages.

### Modified Capabilities
<!-- No requirement changes to existing specs, as the Go rewrite's scope remains the "Happy Path" vertical slice. -->

## Impact

- `go/internal/security/nas/context.go`: `SecurityContext` state and protection logic.
- `go/internal/rrc/builder.go`: RRC message construction logic.
- `go/internal/ue/nas.go`: Integration of stateful security context.
- Interoperability with 3GPP-compliant gNodeB and Core Network implementations.
