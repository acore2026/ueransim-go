# UERANSIM Go Rewrite TODO List

This file tracks the progress of migrating UERANSIM from C++ to Go.

## 1. Foundational Encoders (The "Language" of 5G)
- [x] **ASN.1 (NGAP/RRC):** Go-native ASN.1 implementation (NGAP via free5gc, RRC simplified).
- [x] **NAS (5GMM/5GSM):** Non-Access Stratum protocol codec.
- [x] **Octet/Bit Utilities:** Port `bit_string` and `bit_buffer` for low-level packet construction.

## 2. Security & Crypto (The "Trust")
- [x] **5G-AKA:** Authentication and Key Agreement procedure.
- [x] **Integrity & Ciphering:** Implement NIA2/NEA2 (Snow3G/AES) for secure Control Plane traffic.
- [x] **Key Derivation:** Implement 3GPP KDF (K_AMF, K_gNB, etc.).

## 3. Transport & Radio Simulation (The "Wire")
- [x] **UDP Transport:** Simulated radio interface and GTP-U transport.
- [x] **SCTP Transport:** Reliable transport for NGAP (AMF connection).
- [x] **TUN Interface:** Virtual network device for UE User Plane.
- [x] **GTP-U:** GPRS Tunnelling Protocol for user-plane data flow.
- [x] **Radio over UDP:** Simulated 5G-NR physical layer implementation (RLS).

## 4. Core Protocol Logic (The "Brain")
- [x] **UE Node:** State machine for UE Registration and PDU Session.
- [x] **gNB Node:** SCTP bridging and radio interface management.
- [x] **RRC Layer:** Radio Resource Control (Simplified bit-packer).
- [x] **NGAP Layer:** Next Generation Application Protocol logic via free5gc.

## 5. Integration & Bootstrap
- [x] **Task Runtime:** Core message-passing and task management system.
- [x] **Logging:** Structured logging for Go binaries.
- [x] **Config Loading:** YAML configuration mapping.
- [x] **CLI Interface:** Interactive CLI for UE and gNB control.

---
*Last Updated: 2026-03-13*

