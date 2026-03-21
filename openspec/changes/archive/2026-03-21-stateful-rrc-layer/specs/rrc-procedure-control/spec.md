## ADDED Requirements

### Requirement: RRC Connection Procedure Control
The RRC layer SHALL control the connection establishment procedure including retransmission and timeouts.

#### Scenario: T300 Timer Expiry
- **WHEN** the RRC layer is in `StateConnecting` and the T300 timer expires before `RRCSetup` is received
- **THEN** the system SHALL transition back to `StateIdle` and signal a connection failure to the upper layers

### Requirement: RRC Reconfiguration Handling
The RRC layer SHALL handle reconfiguration messages from the gNB to update radio parameters while in `StateConnected`.

#### Scenario: Successful Reconfiguration
- **WHEN** the RRC layer is in `StateConnected` and receives an `RRCReconfiguration` message
- **THEN** the system SHALL process the message and send an `RRCReconfigurationComplete` response
