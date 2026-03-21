## Why

The current Go implementation of UERANSIM bypasses the Radio Link Control (RLC) layer by sending RRC and user-plane data directly over the Radio Link Simulation (RLS) interface. To achieve full 3GPP protocol stack depth and support features like fragmentation, reassembly, and different transport modes, a native Go implementation of the RLC layer is required. Implementing RLC Transparent Mode (TM) and Unacknowledged Mode (UM) is the first step towards this goal.

## What Changes

- **RLC Layer**: Create a new `go/internal/rlc` package.
- **RLC-TM**: Implement Transparent Mode for RRC CCCH messages (e.g., `RRCSetupRequest`).
- **RLC-UM**: Implement Unacknowledged Mode for RRC DCCH messages and user-plane data.
- **Stack Plumbing**: Integrate the RLC layer between RRC/User-Plane and the RLS layer in both UE and gNB.
- **Framing**: Implement RLC PDU headers and basic segmentation/reassembly logic for UM.

## Capabilities

### New Capabilities
- `rlc-transparent-mode`: Support for RLC-TM as defined in 3GPP TS 38.322.
- `rlc-unacknowledged-mode`: Support for RLC-UM including basic framing and reassembly.

### Modified Capabilities
- `rrc-interop-packing`: Update requirements to ensure RRC messages are passed through the RLC layer before simulation.

## Impact

- `go/internal/rlc/`: (New package) RLC entities and PDU codecs.
- `go/internal/ue/node.go`: Plumbing of the new RLC task.
- `go/internal/gnb/node.go`: Plumbing of the new RLC task in gNB.
- `go/internal/ue/rrc.go` & `go/internal/ue/nas.go`: Update to send data via RLC.
