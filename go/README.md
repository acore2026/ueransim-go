# Go Rewrite Implementation

This subtree contains the Go implementation of UERANSIM, migrated from the original C++ codebase.

## Current Scope

- **Supported Happy Path:** One readable vertical slice from UE bootstrap through registration, security activation, registration completion, and initial PDU session establishment against live free5GC containers.
- **Core Runtime:** Shared task-based actor model with structured logging.
- **Transport Layer:** 
    - Full SCTP support for NGAP/N2 interface.
    - UDP transport for simulated radio and GTP-U.
    - Linux TUN interface for UE User Plane.
- **Protocol Codecs:**
    - **NGAP:** Fully supported via free5gc/ngap.
- **NAS:** 5GMM/5GSM encoding/decoding foundation (Registration, etc.).
    - **RRC:** Initial UPER-compatible bit-packing for connection setup.
    - **GTP-U:** User plane tunneling header implementation.
    - **RLS:** Radio Link Simulation protocol for UDP-based radio.
- **Security & Crypto:** 
    - 5G-AKA (Milenage) authentication.
    - NAS Integrity (NIA2) and Ciphering (NEA2) using AES.
    - 3GPP Key Derivation (KDF).
- **Nodes:** Functional UE and gNodeB state machines for live registration and initial session-establishment signaling with free5GC.
- **NAS Transport:** Bidirectional NAS PDU relaying between UE and Core (Uplink/Downlink), including the narrow happy-path PDU session trigger and NAS delivery from `InitialContextSetupRequest`.
- **NAS Adapter Boundary:** The local `go/internal/nas` package now acts as a stable adapter over `github.com/acore2026/nas` for the supported happy-path messages. UE and gNB procedure code remain handwritten and continue to consume local helper functions and DTOs instead of external NAS package types directly.
- **Bootstrap:** YAML configuration loading and basic interactive CLI.
- **Documentation:** Added `go/VERIFY_REGISTRATION.md` for end-to-end testing.

## In Progress / Future Work

- **User-Plane Readiness:** The current slice reaches `PDU Session Establishment Accept`, but the full bearer/context completion needed for verified user-plane traffic is still incomplete.
- **Full Feature Parity:** The C++ implementation contains many specialized IEs, handover scenarios, and edge-case protocol behaviors that are being incrementally ported.
- **Advanced RRC:** Transitioning from the simplified bit-packer to a full UPER codec.
- **Performance Tuning:** Optimization of the Go runtime for high-throughput User Plane traffic.
- **Comprehensive Integration Tests:** End-to-end testing against 5G Core implementations (Open5GS, free5GC).

## Explicit Non-Goals For The Current Slice

- Full C++ parity across registration, RRC, and session-management edge cases.
- Handover, recovery branches, multi-session orchestration, and parity-only helper layers.
- Generated procedure logic. The Go rewrite keeps state progression handwritten and uses generated or schema-driven helpers only for repetitive protocol structure handling.
- Leaking external NAS message or IE types into the UE/gNB procedure layer. Standard NAS structure handling belongs in the local adapter boundary, not in handwritten procedure code.
