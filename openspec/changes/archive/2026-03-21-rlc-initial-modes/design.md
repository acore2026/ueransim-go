## Context

The current architecture lacks an RLC layer, which is responsible for framing, segmentation, and error handling in the 3GPP stack. Adding RLC Transparent Mode (TM) and Unacknowledged Mode (UM) provides the necessary infrastructure for more advanced protocol features and better adherence to the 5G standards.

## Goals / Non-Goals

**Goals:**
- Implement a task-based RLC layer for both UE and gNB.
- Support RLC-TM for CCCH messages (pass-through).
- Support RLC-UM for DCCH and DTCH (basic framing with sequence numbers).
- Update the stack plumbing to include the RLC layer.

**Non-Goals:**
- RLC Acknowledged Mode (AM) with ARQ/Retransmissions.
- Complex segmentation and reassembly (initial focus on simple framing).
- Multiple RLC entities per UE (start with one for control plane and one for user plane).

## Decisions

### 1. RLC Task Architecture
- **Decision**: Implement RLC as a separate `runtime.Task` that sits between the upper layers (RRC/User-Plane) and the lower layer (RLS).
- **Rationale**: Matches the modular, actor-based design of the Go rewrite. Allows independent state management for RLC entities.

### 2. Mode Mapping
- **CCCH (Common Control Channel)** -> RLC-TM. Used for `RRCSetupRequest` and `RRCSetup`.
- **DCCH (Dedicated Control Channel)** -> RLC-UM. Used for `RRCSetupComplete` and subsequent NAS transport.
- **DTCH (Dedicated Traffic Channel)** -> RLC-UM. Used for user-plane data from TUN.

### 3. UM Framing (6-bit SN)
- **Decision**: Use a simplified RLC-UM header with a 2-bit Segmentation Information (SI) and 6-bit Sequence Number (SN).
- **Header Byte**: `[SI (2 bits)][SN (6 bits)]`.
- **SI = 00**: Complete SDU (no segmentation).

### 4. RLS Protocol Update
- **Decision**: RLS will now carry RLC PDUs instead of raw RRC/NAS PDUs.
- **Rationale**: Correctly mirrors the standard where RRC PDUs are encapsulated in RLC.

## Risks / Trade-offs

- **[Risk] Increased Latency** → Adding a new task in the path might slightly increase internal processing time.
- **[Mitigation]** → Use efficient channel communication and minimize allocations in the RLC task.
- **[Risk] Complexity in Reassembly** → If packets arrive out of order or segmented.
- **[Mitigation]** → Initial implementation will assume no segmentation (SI=00) and focus on framing.
- **[Risk] Breaking gNB/UE Interop** → UE and gNB must both be updated simultaneously to understand the new RLC framing.
- **[Mitigation]** → Update both components in the same change and verify against each other.
