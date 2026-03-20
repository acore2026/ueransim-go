## 1. Adapter Boundary

- [x] 1.1 Inventory the currently used happy-path message builders and decoders in `go/internal/nas` and map each one to a `github.com/acore2026/nas` equivalent.
- [x] 1.2 Add the forked NAS module to the Go module wiring for local development and tests.
- [x] 1.3 Define the local adapter entry points and local DTOs that preserve the current procedure-layer interfaces.

## 2. Codec Migration

- [x] 2.1 Migrate registration, authentication, identity, and security-mode message handling in `go/internal/nas` to the adapter-backed codec path.
- [x] 2.2 Migrate `Registration Complete`, `UL NAS Transport`, `DL NAS Transport`, and initial PDU session establishment message handling to the adapter-backed codec path.
- [x] 2.3 Remove or retire the replaced manual codec paths once the adapter-backed implementations are verified.

## 3. Verification

- [x] 3.1 Update NAS-focused unit tests to validate the adapter-backed encode/decode path for the supported happy-path messages.
- [x] 3.2 Re-run the current Go happy-path registration and initial PDU-session verification against the live free5GC containers.
- [x] 3.3 Document the adapter boundary so future procedure work continues to use local helpers rather than external NAS types directly.
