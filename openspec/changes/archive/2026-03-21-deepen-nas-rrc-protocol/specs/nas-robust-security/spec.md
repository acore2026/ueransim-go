## ADDED Requirements

### Requirement: 32-bit COUNT Management
The NAS security context SHALL maintain a 32-bit COUNT value for both uplink and downlink, consisting of a 24-bit overflow counter and an 8-bit sequence number (SQN).

#### Scenario: Downlink COUNT reconstruction with SQN wrap-around
- **WHEN** the current downlink SQN is 255 and a NAS message with SQN 0 is received
- **THEN** the system SHALL increment the 24-bit overflow counter and set the 32-bit COUNT accordingly

#### Scenario: Uplink COUNT increment on transmission
- **WHEN** a NAS message is sent with security protection
- **THEN** the system SHALL increment the uplink SQN, and if it wraps around to 0, increment the 24-bit overflow counter

### Requirement: NAS Replay Protection
The system SHALL maintain a sliding window of recently received NAS sequence numbers to detect and reject replayed messages.

#### Scenario: Duplicate sequence number rejection
- **WHEN** a protected NAS message is received with a sequence number already present in the 16-message sliding window
- **THEN** the system SHALL reject the message with a "MAC verification failed" error or equivalent security failure
