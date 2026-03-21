## ADDED Requirements

### Requirement: PDU Session Release Request
The UE SHALL be able to initiate a PDU Session Release procedure by sending a `PDU Session Release Request` to the network.

#### Scenario: UE-initiated release
- **WHEN** the UE is in `StatePduSessionActive` and a release is triggered
- **THEN** the system SHALL send a `PDU Session Release Request` with a new PTI

### Requirement: PDU Session Release Command Handling
The UE SHALL process `PDU Session Release Command` messages from the SMF and respond with `PDU Session Release Complete`.

#### Scenario: Successful release command
- **WHEN** the UE receives a valid `PDU Session Release Command`
- **THEN** the system SHALL send a `PDU Session Release Complete` response and trigger user-plane cleanup

### Requirement: User Plane Cleanup on Release
The system SHALL stop the corresponding user-plane tasks and remove the TUN device when a PDU session is released.

#### Scenario: Cleanup after release
- **WHEN** a PDU session release procedure is completed
- **THEN** the system SHALL stop the `ue-tun` task and remove the `uesimtun` device
