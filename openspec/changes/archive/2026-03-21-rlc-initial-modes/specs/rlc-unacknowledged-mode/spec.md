## ADDED Requirements

### Requirement: RLC-UM Sequence Numbering
The RLC-UM entity SHALL maintain an independent 6-bit sequence number counter for transmission.

#### Scenario: Sequence number increment
- **WHEN** multiple SDUs are sent via the same RLC-UM entity
- **THEN** each successive PDU SHALL have a sequence number incremented by 1 (modulo 64)

### Requirement: RLC Stack Integration
The RLC layer SHALL be correctly plumbed between the upper layers (RRC/User-Plane) and the simulation layer (RLS).

#### Scenario: End-to-end UE-gNB communication
- **WHEN** the Go UE sends a message to the Go gNB via the new RLC-enabled stack
- **THEN** both ends SHALL correctly encapsulate and decapsulate the PDUs using the agreed RLC modes
