# Go Rewrite Implementation

This subtree contains the Go implementation of UERANSIM, migrated from the original C++ codebase.

## Current Scope

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
- **Nodes:** Functional UE and gNodeB state machines for initial registration.
- **Bootstrap:** YAML configuration loading and basic interactive CLI.

## In Progress / Future Work

- **Full Feature Parity:** The C++ implementation contains many specialized IEs, handover scenarios, and edge-case protocol behaviors that are being incrementally ported.
- **Advanced RRC:** Transitioning from the simplified bit-packer to a full UPER codec.
- **Performance Tuning:** Optimization of the Go runtime for high-throughput User Plane traffic.
- **Comprehensive Integration Tests:** End-to-end testing against 5G Core implementations (Open5GS, free5GC).
