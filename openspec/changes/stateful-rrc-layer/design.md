## Context

The current `RrcTaskHandler` in `go/internal/ue/rrc.go` is a simple relay. It wraps every incoming NAS PDU in an RRC container and sends it to RLS, assuming the connection is always available. In a real 5G stack, the UE must explicitly request a connection (`RRCSetupRequest`), wait for a setup message (`RRCSetup`), and then signal completion (`RRCSetupComplete`).

## Goals / Non-Goals

**Goals:**
- Implement a state-driven `RrcTaskHandler` with support for `RRC-IDLE`, `RRC-CONNECTING`, and `RRC-CONNECTED`.
- Handle the full 3-way handshake for RRC connection establishment.
- Buffering NAS messages that arrive while the RRC connection is being established.
- Support `RRCReconfiguration` and `RRCRelease` for basic lifecycle management.

**Non-Goals:**
- Full UPER decoding of all RRC messages (we will continue to use targeted bit-extraction for specific fields).
- Support for complex handover or dual-connectivity scenarios.
- Radio resource configuration (e.g., physical layer parameters) beyond the vertical slice needs.

## Decisions

### 1. RRC State Machine
- **Decision**: Define a formal `RrcState` enum and handle transitions within the `OnMessage` loop of the RRC task.
- **Rationale**: Ensures that the RRC layer behaves predictably and only sends data when the connection is active.
- **States**:
    - `StateIdle`: No connection. Triggered by start or `RRCRelease`.
    - `StateConnecting`: Sent `RRCSetupRequest`, waiting for `RRCSetup`.
    - `StateConnected`: Received `RRCSetup`, sent `RRCSetupComplete`. Ready for NAS transport.

### 2. NAS Message Buffering
- **Decision**: When NAS sends a PDU while RRC is in `StateIdle`, RRC will transition to `StateConnecting`, send `RRCSetupRequest`, and buffer the NAS PDU in an internal queue.
- **Rationale**: NAS should not need to care about the RRC connection state for initial registration. RRC handles the "dial-up" logic.

### 3. T300 Timer Implementation
- **Decision**: Use a Go timer or a task-based timeout message to implement the T300 timer (Wait for `RRCSetup`).
- **Rationale**: Standard 3GPP requirement to prevent hanging in the connecting state if the gNB does not respond.

### 4. Downlink Decoding
- **Decision**: Use the same bit-matching approach as implemented in the gNB to detect `RRCSetup` and `RRCRelease` from the RLS stream.
- **Rationale**: Avoids the complexity of a full ASN.1 decoder while maintaining bit-accuracy for the vertical slice.

## Risks / Trade-offs

- **[Risk] State Desync** → The UE might think it's connected while the gNB has timed out the context.
- **[Mitigation]** → Handle `RRCRelease` and implement basic "Inactive" or "Idle" transitions if no traffic is seen for a long period (future work).
- **[Risk] Complex Bit-Extraction** → Downlink messages (DL-CCCH/DL-DCCH) have different bit-layouts.
- **[Mitigation]** → Carefully verify the bit-indices for each message against TS 38.331 and the C++ reference.
