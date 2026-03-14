# UERANSIM Go Rewrite TODO List

This file tracks the progress of migrating UERANSIM from C++ to Go.

## 1. Foundational Encoders (The "Language" of 5G)
- [x] **ASN.1 (NGAP/RRC):** Go-native ASN.1 implementation (NGAP via free5gc, RRC simplified).
- [x] **NAS (5GMM/5GSM):** Non-Access Stratum protocol codec.
- [x] **Octet/Bit Utilities:** Port `bit_string` and `bit_buffer` for low-level packet construction.

## 2. Security & Crypto (The "Trust")
- [x] **5G-AKA:** Authentication and Key Agreement procedure (Handshake verified).
- [ ] **NAS Security Context:** Implement stateful security context (Sequence numbers, integrity/ciphering keys).
- [x] **Integrity & Ciphering:** Implement NIA2/NEA2 (AES-CMAC/AES-CTR).
- [x] **Key Derivation:** Implement 3GPP KDF (K_AMF, K_gNB, K_NAS_int, K_NAS_enc).

## 3. Transport & Radio Simulation (The "Wire")
- [x] **UDP Transport:** Simulated radio interface and GTP-U transport.
- [x] **SCTP Transport:** Reliable transport for NGAP (AMF connection).
- [x] **TUN Interface:** Virtual network device for UE User Plane.
- [x] **GTP-U:** GPRS Tunnelling Protocol for user-plane data flow.
- [x] **Radio over UDP:** Bidirectional relaying of RLS packets between UE and gNB.

## 4. Core Protocol Logic (The "Brain")
- [ ] **UE State Machine:** Complete Registration flow (Security Mode, Registration Accept).
- [ ] **PDU Session:** State machine for PDU Session Establishment and Release.
- [x] **gNB Node:** SCTP bridging and radio interface management.
- [x] **RRC Layer:** Radio Resource Control (Bidirectional container relaying).
- [x] **NGAP Layer:** Next Generation Application Protocol logic (InitialUEMessage, UplinkNASTransport).

## 5. Integration & Bootstrap
- [x] **Task Runtime:** Core message-passing and task management system.
- [x] **Logging:** Structured logging for Go binaries.
- [x] **Config Loading:** YAML configuration mapping.
- [x] **CLI Interface:** Interactive CLI for UE and gNB control.

---
*Last Updated: 2026-03-14*
