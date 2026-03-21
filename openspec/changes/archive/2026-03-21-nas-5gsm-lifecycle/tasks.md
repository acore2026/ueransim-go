## 1. Codec Refactoring & Expansion

- [x] 1.1 Create `go/internal/nas/msg_sm.go` and move existing SM messages from `msg_mm.go`.
- [x] 1.2 Add new 5GSM message types to `go/internal/nas/nas.go`.
- [x] 1.3 Implement `PDUSessionModificationRequest`, `Command` (decode), and `Complete` in `msg_sm.go`.
- [x] 1.4 Implement `PDUSessionReleaseRequest`, `Command` (decode), and `Complete` in `msg_sm.go`.

## 2. NAS SM Procedure Logic

- [x] 2.1 Implement PTI management (counter and lookup) in `NasTaskHandler`.
- [x] 2.2 Update `handleDlNasTransport` to dispatch 5GSM messages to new handlers.
- [x] 2.3 Implement `handlePduSessionModificationCommand` and `handlePduSessionReleaseCommand`.
- [x] 2.4 Implement `sendPduSessionModificationRequest` and `sendPduSessionReleaseRequest`.

## 3. User Plane Cleanup

- [x] 3.1 Implement a mechanism to stop `ue-tun` task from `NasTaskHandler` upon session release.
- [x] 3.2 Ensure the TUN device is correctly closed and removed.

## 4. Verification & Testing

- [x] 4.1 Add CLI commands to trigger PDU Session Modification and Release.
- [x] 4.2 Verify the full Modification and Release flows against live free5GC core using logs.
- [x] 4.3 Confirm resource cleanup (TUN device removal) after session release.
- [x] 4.4 Commit and push changes.
