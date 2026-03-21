## Why

The current Go implementation of UERANSIM only supports the initial PDU Session Establishment procedure. To achieve protocol depth and support dynamic session management, the implementation must include the PDU Session Modification and PDU Session Release procedures. This allows the UE to update its session parameters (e.g., QoS, slice info) and gracefully clean up resources when they are no longer needed.

## What Changes

- **NAS SM Codec**: Implement encoders and decoders for the following 5GSM messages:
    - `PDU Session Modification Request/Command/Complete`
    - `PDU Session Release Request/Command/Complete`
- **NAS SM State Machine**: Update the UE's NAS task to handle session modification and release flows.
- **CLI Triggering**: Add interactive CLI commands to the Go UE to manually trigger session modification and release.
- **Resource Cleanup**: Ensure that PDU Session Release correctly triggers the removal of the corresponding TUN device or route configuration.

## Capabilities

### New Capabilities
- `nas-5gsm-modification`: Full PDU Session Modification procedure (UE-initiated and network-solicited).
- `nas-5gsm-release`: Full PDU Session Release procedure (UE-initiated and network-solicited).

### Modified Capabilities
<!-- No requirement changes to existing specs -->

## Impact

- `go/internal/nas/nas.go`: Definition of new 5GSM message types.
- `go/internal/nas/msg_sm.go`: (New file) Codecs for 5GSM lifecycle messages.
- `go/internal/ue/nas.go`: Logic for handling Modification/Release commands and sending requests.
- `go/internal/core/cli`: New commands for UE control.
