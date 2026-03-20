## ADDED Requirements

### Requirement: Adapter-backed NAS happy-path codecs
The system SHALL encode and decode the supported happy-path NAS messages through a local adapter backed by the forked `github.com/acore2026/nas` module instead of relying on manual byte-level message implementations in the procedure path.

#### Scenario: Encode registration request through adapter
- **WHEN** the UE NAS task builds a supported `Registration Request`
- **THEN** the local NAS layer SHALL use the adapter-backed codec path to produce the NAS PDU
- **AND** the procedure layer SHALL not need to manipulate `free5gc/nas` message or IE structs directly

#### Scenario: Decode downlink happy-path NAS messages through adapter
- **WHEN** the UE receives a supported downlink NAS message in the current happy path
- **THEN** the local NAS layer SHALL decode it through the adapter-backed codec path
- **AND** the procedure layer SHALL receive only the fields required for its state transition decisions

### Requirement: Preserve readable handwritten procedure logic
The system SHALL preserve handwritten UE and gNB procedure logic as the owner of state progression and SHALL keep external NAS package types behind the local NAS adapter boundary.

#### Scenario: Procedure code consumes local helpers
- **WHEN** the UE NAS task handles authentication, identity, security mode, registration accept, or PDU session signaling
- **THEN** it SHALL continue to call local helper functions or consume local result types
- **AND** it SHALL not depend on `nasMessage` or `nasType` objects from the external fork

### Requirement: Preserve current happy-path behavior
The system SHALL preserve the current supported registration and initial PDU-session signaling behavior while migrating the NAS codec layer.

#### Scenario: Existing happy-path verification remains valid
- **WHEN** the adapter-backed NAS codec path is used for the supported registration and initial PDU-session flow
- **THEN** the UE SHALL still complete registration successfully
- **AND** the UE SHALL still send `PDU Session Establishment Request`
- **AND** the UE SHALL still receive `PDU Session Establishment Accept`

### Requirement: Use the local NAS fork as the migration dependency
The system SHALL use the locally available forked NAS module during migration so that codec gaps found in the supported happy path can be corrected without blocking the rewrite.

#### Scenario: Fork-backed dependency is wired for migration
- **WHEN** the NAS adapter is introduced
- **THEN** the project SHALL resolve the NAS codec dependency to the forked module at `/root/proj/go/acore-forks/nas`
- **AND** the adapter design SHALL allow targeted fixes in the fork without changing procedure-layer interfaces
