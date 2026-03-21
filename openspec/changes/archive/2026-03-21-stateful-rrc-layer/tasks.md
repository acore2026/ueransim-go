## 1. RRC State Machine Scaffolding

- [x] 1.1 Define `RrcState` enum in `go/internal/ue/rrc.go`.
- [x] 1.2 Update `RrcTaskHandler` to include `state` and a buffer for pending NAS messages.
- [x] 1.3 Implement `T300` timer management in `RrcTaskHandler`.

## 2. Authentic Bit-Packing Expansion

- [x] 2.1 Implement `BuildRRCReconfigurationComplete` in `go/internal/rrc/builder.go`.
- [x] 2.2 Add unit tests for the new bit-packer in `go/internal/rrc/builder_test.go`.

## 3. RRC Procedure Implementation

- [x] 3.1 Implement `handleNasToRrc` logic to trigger connection establishment if in `StateIdle`.
- [x] 3.2 Implement `handleRlsToRrc` logic to decode `RRCSetup`, `RRCReconfiguration`, and `RRCRelease`.
- [x] 3.3 Implement state transitions and draining of buffered NAS messages upon successful setup.

## 4. Verification & Cleanup

- [x] 4.1 Verify state transitions using logs during a registration run against free5GC.
- [x] 4.2 Verify the `T300` timer by simulating a gNB non-response.
- [x] 4.3 Ensure no regressions in user-plane traffic.
- [x] 4.4 Commit and push changes.
