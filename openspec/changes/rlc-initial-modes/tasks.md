## 1. RLC Package & Codecs

- [x] 1.1 Create `go/internal/rlc/` package and define `Mode` and `Header` structures.
- [x] 1.2 Implement RLC-UM header packing and unpacking (1-byte header, 6-bit SN).
- [x] 1.3 Add unit tests for RLC-UM framing in `go/internal/rlc/rlc_test.go`.

## 2. RLC Task Implementation

- [x] 2.1 Implement `RlcTaskHandler` in `go/internal/rlc/task.go`.
- [x] 2.2 Handle `upper_to_rlc` message (encapsulate based on mode).
- [x] 2.3 Handle `rls_to_rlc` message (decapsulate and deliver to RRC/NAS).

## 3. UE Stack Plumbing

- [x] 3.1 Update `go/internal/ue/node.go` to instantiate and plumb the RLC task.
- [x] 3.2 Update `go/internal/ue/rrc.go` to send/receive via RLC.
- [x] 3.3 Update `go/internal/ue/rls.go` to deliver packets to RLC instead of RRC directly.

## 4. gNB Stack Plumbing & Verification

- [x] 4.1 Update `go/internal/gnb/node.go` (if exists) or equivalent to include RLC.
- [x] 4.2 Update `go/internal/gnb/rls.go` to handle RLC-encapsulated PDUs.
- [x] 4.3 Verify full registration and PDU session flow with the new RLC layer.
- [x] 4.4 Commit and push changes.
