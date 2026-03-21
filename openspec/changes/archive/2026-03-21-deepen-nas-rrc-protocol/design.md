## Context

The current Go implementation for UERANSIM is a vertical slice that achieves a "Happy Path" registration. However, it lacks protocol depth in two critical areas:
1. **NAS Security**: The 32-bit COUNT is treated as a simple 8-bit sequence number without handling rollovers or maintaining the 24-bit overflow counter. This leads to cryptographic failure after 256 messages.
2. **RRC Interoperability**: Instead of using 3GPP-standard Unaligned Packed Encoding Rules (UPER), the Go version uses a custom "SimpleRRC" TLV wrapper. This prevents the Go UE from communicating with any 3GPP-compliant gNodeB (including the original UERANSIM C++ gNB).

## Goals / Non-Goals

**Goals:**
- Implement 3GPP-compliant 32-bit COUNT management in the NAS security context.
- Add replay protection for NAS messages using a sequence number sliding window.
- Replace the custom `SimpleRRC` with a bit-accurate UPER-compliant `RRCSetupComplete` message.
- Maintain the existing registration flow success against free5GC.

**Non-Goals:**
- Implementation of a generalized ASN.1 UPER compiler/library.
- Full protocol parity with the C++ version (Handover, RLC, etc. remain out of scope).
- Supporting all RRC messages (focus only on the UL-DCCH path for registration).

## Decisions

### 1. NAS Count Reconstruction
- **Decision**: Mirror the C++ `NasCount` architecture using a 24-bit overflow and 8-bit sequence number (SQN).
- **Rationale**: This is the standard 3GPP approach. The 32-bit COUNT is required for the AES-CTR (ciphering) and AES-CMAC (integrity) algorithms.
- **Logic**: When an 8-bit SN is received, the 32-bit COUNT is reconstructed by comparing the received SN with the local state. If `SN_rx < SN_local`, a rollover is assumed, and the overflow counter is incremented.

### 2. Replay Protection Window
- **Decision**: Implement a 16-entry sliding window for received NAS sequence numbers.
- **Rationale**: Matches the original C++ implementation's security posture. Any message with a sequence number already present in the "last received" window is rejected as a replay attack.

### 3. Manual UPER Bit-Packing for RRC
- **Decision**: Hand-code the bit-level structure of `RRCSetupComplete` using the `utils.BitString` utility.
- **Rationale**: A full ASN.1 UPER library is too heavy for the current slice. Manual packing ensures 3GPP compatibility for the specific messages we need while keeping the codebase lean.
- **Mapping**: The bit-stream will follow TS 38.331 for the `UL-DCCH-Message` choice and the `RRCSetupComplete` sequence fields.

## Risks / Trade-offs

- **[Risk] Bit-packing Errors** → Manual bit-packing is error-prone and hard to debug.
- **[Mitigation]** → Use Wireshark to verify the generated RRC packets against the original C++ UERANSIM output.
- **[Risk] SQN Desync** → Aggressive overflow incrementing could desync the UE from the AMF if packets are lost or reordered.
- **[Mitigation]** → Use the standard reconstruction logic that allows for small jumps in SN without triggering overflow until a wrap-around is detected.
