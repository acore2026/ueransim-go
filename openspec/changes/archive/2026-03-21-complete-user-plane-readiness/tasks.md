## 1. Extract and model the bearer setup state

- [x] 1.1 Identify the `InitialContextSetupRequest` parsing path in `go/internal/ngap` and add the helper surface needed to extract PDU session resource setup details for the supported flow.
- [x] 1.2 Introduce explicit gNB-side PDU session state for UE association, remote UPF tunnel info, local tunnel info, and session lookup keys.
- [x] 1.3 Allocate and track the local TEID and GTP bind information needed to complete the supported bearer setup path.

## 2. Complete the control-plane side of user-plane setup

- [x] 2.1 Update the gNB NGAP handling so `InitialContextSetupRequest` processing creates or updates the session record instead of only relaying embedded NAS.
- [x] 2.2 Build and send the required NGAP bearer/setup response after local tunnel state is ready for the supported session.
- [x] 2.3 Add focused tests for NGAP session-resource extraction and gNB session-state creation.

## 3. Build the minimal GTP-U forwarding path

- [x] 3.1 Replace the placeholder gNB GTP heartbeat task with a minimal real GTP-U socket/task for the supported happy path.
- [x] 3.2 Wire UE-originated packets from the existing TUN path into uplink GTP-U encapsulation using the established session mapping.
- [x] 3.3 Implement downlink GTP-U decapsulation and delivery back to the matching UE session path.
- [x] 3.4 Gate user-plane activation so the UE session is not treated as ready until bearer and tunnel setup have completed.

## 4. Verify and document user-plane readiness

- [x] 4.1 Add or update focused tests for session gating and the minimal GTP/session mapping boundaries.
- [x] 4.2 Update the Go verification documentation to include bearer setup, expected user-plane-ready milestones, and any observable packet-flow confirmation steps.
- [x] 4.3 Validate the supported flow against the running NF containers on the machine, starting with `docker ps` and ending with documented evidence of basic user-plane readiness.
