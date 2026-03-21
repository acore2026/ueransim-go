## Context

The current `NasTaskHandler` handles the registration flow and one-time PDU session establishment. Once the session is established, there is no logic to modify or release it. 5G SM procedures are stateful and use Procedure Transaction Identities (PTI) to correlate UE requests with SMF responses.

## Goals / Non-Goals

**Goals:**
- Implement the PDU Session Modification procedure.
- Implement the PDU Session Release procedure.
- Organize SM-specific codecs into a dedicated `go/internal/nas/msg_sm.go` file.
- Add basic PTI management to the NAS task.
- Ensure user-plane cleanup (TUN device) on session release.

**Non-Goals:**
- Multiple concurrent PDU sessions (remain limited to one session for this slice).
- Complex QoS flow management.
- Support for all SM optional IEs.

## Decisions

### 1. Codec Organization
- **Decision**: Create `go/internal/nas/msg_sm.go` and move existing PDU Session Establishment messages there, along with the new Modification/Release messages.
- **Rationale**: MM and SM are distinct sub-layers in 5G NAS. Keeping them separate improves maintainability as the protocol depth increases.

### 2. PTI Management
- **Decision**: Implement a simple incrementing PTI counter in the `NasTaskHandler`.
- **Rationale**: While only one session is supported, correctly using PTI is a standard requirement and prevents protocol errors with the SMF.

### 3. Procedure Handling
- **Decision**: Update `handleDlNasTransport` to dispatch to specific SM message handlers.
- **Modification Flow**: UE sends Request -> SMF sends Command -> UE sends Complete.
- **Release Flow**: UE sends Request -> SMF sends Command -> UE sends Complete.

### 4. User Plane Cleanup
- **Decision**: PDU Session Release will trigger a `Stop` signal to the `ue-tun` task.
- **Rationale**: Ensures resources are freed and the OS network state is cleaned up.

## Risks / Trade-offs

- **[Risk] SMF Unsolicited Messages** → SMF might send a Modification Command without a UE Request.
- **[Mitigation]** → Ensure handlers can handle network-solicited commands (PTI=0).
- **[Risk] State Complexity** → NAS task state machine becoming too large.
- **[Mitigation]** → Use clear transition logic and potentially split SM logic into a helper structure if it gets too complex.
