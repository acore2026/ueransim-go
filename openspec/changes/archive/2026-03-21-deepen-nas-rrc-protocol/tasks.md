## 1. Robust NAS Security Context

- [x] 1.1 Update `SecurityContext` struct in `go/internal/security/nas/context.go` to handle 32-bit COUNT (24-bit overflow + 8-bit SQN).
- [x] 1.2 Implement `NasCount` rollover and reconstruction logic (`estimatedDownlinkCount`) in `context.go`.
- [x] 1.3 Implement a 16-message sliding window for replay protection in `SecurityContext`.
- [x] 1.4 Update `Protect` and `Unprotect` functions to use the full 32-bit COUNT for integrity and ciphering algorithms.
- [x] 1.5 Add unit tests for `NasCount` rollover and replay protection in `go/internal/security/nas/algorithms_test.go` or a new test file.

## 2. Authentic RRC Bit-Packing

- [x] 2.1 Implement authentic 3GPP TS 38.331 UPER bit-packing for `RRCSetupComplete` in `go/internal/rrc/builder.go`.
- [x] 2.2 Remove the non-standard `SimpleRRC` TLV wrapper and associated logic from `builder.go`.
- [x] 2.3 Verify the bit-stream of `RRCSetupComplete` matches the expected 3GPP structure (manual bit-level verification).

## 3. System Integration & Verification

- [x] 3.1 Verify the Go UE can still complete the full registration flow against live free5GC containers.
- [x] 3.2 Confirm `uesimtun0` is configured with the correct IP after PDU Session Establishment.
- [x] 3.3 Verify basic uplink GTP-U packet forwarding from the UE through the gNB.
- [x] 3.4 Perform a final cleanup and ensure no regressions in existing Go components.

## 4. Finalization

- [x] 4.1 Commit all changes with a descriptive message.
- [x] 4.2 Push the changes to the remote repository.
