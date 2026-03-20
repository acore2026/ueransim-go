## Why

The current Go NAS layer manually encodes and decodes a narrow set of 5GMM and 5GSM messages. That kept the first happy path readable, but it will slow down further protocol work and increase wire-format bugs as more NAS messages and IEs are added.

## What Changes

- Introduce a local NAS adapter layer that uses the forked `github.com/acore2026/nas` module as the codec substrate for standard NAS message and IE encoding/decoding.
- Keep UE and gNB procedure logic handwritten and readable by exposing small local helper functions and DTOs instead of leaking `free5gc/nas` types into upper layers.
- Replace the current manually maintained happy-path NAS message builders/parsers in `go/internal/nas` with adapter-backed implementations for the supported flow.
- Add verification coverage to ensure the new adapter preserves the current successful registration and initial PDU session signaling behavior.

## Capabilities

### New Capabilities
- `nas-codec-adapter`: The system uses a local adapter over the forked `github.com/acore2026/nas` module to encode and decode the supported NAS happy-path messages while preserving readable handwritten procedure logic.

### Modified Capabilities

## Impact

- Affected code: `go/internal/nas`, `go/internal/ue/nas.go`, NAS-related tests, and module dependency wiring.
- New dependency usage: local fork at `/root/proj/go/acore-forks/nas` with module path `github.com/acore2026/nas`.
- Verification impact: existing registration and PDU-session happy-path tests must continue to pass against the adapted codec layer.
