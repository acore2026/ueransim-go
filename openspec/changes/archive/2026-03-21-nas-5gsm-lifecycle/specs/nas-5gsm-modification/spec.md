## ADDED Requirements

### Requirement: PDU Session Modification Request
The UE SHALL be able to initiate a PDU Session Modification procedure by sending a `PDU Session Modification Request` to the network.

#### Scenario: UE-initiated modification
- **WHEN** the UE is in `StatePduSessionActive` and a modification is triggered
- **THEN** the system SHALL send a `PDU Session Modification Request` with a new PTI

### Requirement: PDU Session Modification Command Handling
The UE SHALL process `PDU Session Modification Command` messages from the SMF and respond with `PDU Session Modification Complete`.

#### Scenario: Successful modification command
- **WHEN** the UE receives a valid `PDU Session Modification Command`
- **THEN** the system SHALL update the session parameters and send a `PDU Session Modification Complete` response
