## ADDED Requirements

### Requirement: RRC State Machine Management
The RRC layer SHALL maintain a formal state machine consisting of at least `StateIdle`, `StateConnecting`, and `StateConnected`.

#### Scenario: Transition to Connecting on NAS PDU
- **WHEN** the RRC layer is in `StateIdle` and receives a NAS PDU from the NAS layer
- **THEN** the system SHALL transition to `StateConnecting` and send an `RRCSetupRequest` to the gNB

#### Scenario: Transition to Connected on RRCSetup
- **WHEN** the RRC layer is in `StateConnecting` and receives a valid `RRCSetup` message
- **THEN** the system SHALL transition to `StateConnected`, send `RRCSetupComplete`, and drain any buffered NAS messages

#### Scenario: Transition to Idle on RRCRelease
- **WHEN** the RRC layer is in `StateConnected` and receives a valid `RRCRelease` message
- **THEN** the system SHALL transition to `StateIdle` and clear all connection-specific context
