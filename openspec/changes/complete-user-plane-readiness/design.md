## Context

The Go rewrite already completes the control-plane portion of the supported happy path: NG setup, initial registration, NAS authentication and security, registration completion, and receipt of `PDU Session Establishment Accept`. The remaining gap is the handoff from control-plane success to a usable data plane.

Today the gNB extracts NAS from `InitialContextSetupRequest`, but it does not complete the PDU session resource setup carried alongside that NAS payload. The current Go tree also has a placeholder gNB GTP task and an existing UE TUN implementation that are not yet assembled into a working `TUN <-> GTP-U <-> UPF` path. This change finishes that narrow path without expanding into broad simulator parity.

Constraints:
- The C++ implementation remains available for broader behavior and is not the target for parity in this change.
- The supported environment is the existing set of Docker-hosted free5GC NFs already running on the machine.
- Procedure orchestration should remain readable and handwritten; repetitive protocol structure handling can live behind helpers.

## Goals / Non-Goals

**Goals:**
- Complete the supported happy path so a Go UE session becomes minimally usable after `PDU Session Establishment Accept`.
- Parse the NGAP session resource content needed to learn UPF tunnel parameters from `InitialContextSetupRequest`.
- Create explicit gNB-side PDU session state for UE association, TEIDs, addresses, and flow/session lookup.
- Replace the placeholder gNB GTP heartbeat with a minimal real GTP-U data path.
- Wire UE TUN packets into the established user plane and deliver downlink packets back to the UE.
- Send the required NGAP setup response for the supported bearer/session setup path.
- Document and verify the resulting path against the live NF environment.

**Non-Goals:**
- Full bearer-management parity with the C++ implementation.
- Multi-session orchestration, handover, release recovery, or advanced QoS policy handling.
- Generalizing the data-plane architecture for every future RAN scenario.
- Supporting interoperability claims beyond the current live NF environment.

## Decisions

### Decision: Extend the existing happy-path capability instead of introducing a new product boundary

This change modifies the existing `go-happy-path-session-flow` capability. The missing work is not a separate feature; it is the unfinished bottom half of the already-supported happy path.

Alternatives considered:
- Create a separate `go-user-plane` capability. Rejected because it would split one end-to-end flow across multiple specs and blur the definition of done.

### Decision: Parse bearer/session setup at the NGAP boundary and hand a reduced session description to the rest of the gNB

The gNB NGAP layer should extract the PDU session resource setup content from `InitialContextSetupRequest` and convert it into a small internal description that includes:
- UE/session association identifiers
- remote UPF GTP-U address
- remote UPF TEID
- supported QoS flow identifiers required by the happy path

This keeps protocol-specific transfer parsing close to NGAP while preventing low-level IE details from leaking into the wider procedure code.

Alternatives considered:
- Parse NGAP transfer content inside the GTP task. Rejected because it couples data-plane plumbing to control-plane wire structures.
- Store raw transfer blobs and decode lazily later. Rejected because it obscures failures and spreads protocol knowledge across modules.

### Decision: Introduce explicit gNB session state as the single source of truth for user-plane mappings

The gNB should create a real session record keyed by UE association and PDU session identity. That record should own:
- remote tunnel endpoint info learned from NGAP
- locally allocated TEID and local GTP bind info
- mapping needed for uplink lookup from UE packets
- mapping needed for downlink lookup from incoming GTP packets

Alternatives considered:
- Reuse ad hoc identifiers already passed around in NGAP handlers. Rejected because user-plane readiness requires durable cross-task state.
- Push session state into the UE-side code. Rejected because the gNB owns the RAN-side bearer/tunnel completion.

### Decision: Keep the initial data plane minimal and tied to one verified happy path

The GTP task should do only the minimum required to prove a usable session:
- bind a real UDP socket for GTP-U
- encapsulate UE-originated packets into GTP-U using the established session mapping
- decapsulate downlink GTP-U packets and route them back to the matching UE path

No attempt will be made in this change to implement advanced bearer lifecycle handling, multiple concurrent session orchestration, or parity-only abstractions from C++.

Alternatives considered:
- Build a generalized multi-bearer framework first. Rejected because it expands scope before the first usable path exists.
- Stop at sending NGAP response without packet forwarding. Rejected because it would still leave the advertised user-plane readiness incomplete.

### Decision: Treat documented packet-flow verification as part of the definition of done

Verification for this change must go beyond NAS milestones. A successful run should demonstrate that:
- the required NF containers are running
- the Go gNB and Go UE complete registration and PDU session establishment
- the session reaches a documented basic user-plane-ready state, with observable evidence from tunnel setup and packet forwarding milestones

Alternatives considered:
- Accept control-plane completion alone. Rejected because that is the already-known incomplete state.

## Risks / Trade-offs

- [InitialContextSetupRequest transfer parsing may pull in complex NGAP details] -> Mitigation: keep parsing confined to NGAP helpers and expose only a reduced internal session description.
- [A minimal GTP-U path may hard-code assumptions that later work outgrows] -> Mitigation: limit the scope explicitly to one verified happy path and keep session state boundaries narrow.
- [Downlink delivery may expose missing UE or gNB concurrency assumptions] -> Mitigation: keep the first implementation single-path-oriented and verify against the existing NF environment before broadening.
- [Live-NF verification can fail for environment reasons unrelated to code] -> Mitigation: keep `docker ps` and expected milestone guidance in the verification flow so failures are easier to localize.

## Migration Plan

1. Update the happy-path spec to require concrete user-plane readiness behavior.
2. Implement NGAP session-resource extraction and gNB session-state creation.
3. Replace the placeholder gNB GTP task with a minimal real GTP-U path and connect it to UE TUN traffic.
4. Add the required NGAP setup response and downlink packet delivery path.
5. Update verification docs to cover packet-flow readiness against the running NF containers.

Rollback remains low-risk because the C++ implementation still provides the broader fallback path.

## Open Questions

- What is the smallest acceptable proof of "user-plane ready" in this repo: tunnel creation only, or at least one documented packet exchange through the TUN/GTP path?
- Does the existing NGAP helper surface already expose enough transfer content, or will this change need a deeper parser boundary for session resource setup items?
- Should UE TUN wiring remain always-on for the supported flow, or only become active after the gNB confirms session setup completion?
