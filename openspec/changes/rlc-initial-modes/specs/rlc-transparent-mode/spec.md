## ADDED Requirements

### Requirement: RLC Transparent Mode
The RLC layer SHALL support Transparent Mode (TM) for CCCH messages, where PDUs are passed to/from lower layers without any modification.

#### Scenario: RRCSetupRequest via RLC-TM
- **WHEN** an `RRCSetupRequest` is sent from RRC to the RLC layer
- **THEN** the RLC layer SHALL forward the PDU to the RLS layer without adding any headers

### Requirement: RLC Unacknowledged Mode Framing
The RLC layer SHALL support Unacknowledged Mode (UM) by adding a 1-byte header containing a 6-bit sequence number.

#### Scenario: NAS Transport via RLC-UM
- **WHEN** an RRC message (DCCH) is sent to the RLC layer
- **THEN** the system SHALL prepend an RLC-UM header with an incrementing sequence number

### Requirement: RLC UM Reassembly
The RLC layer SHALL extract the original SDU from the RLC-UM PDU by removing the header.

#### Scenario: Receiving RLC-UM PDU
- **WHEN** an RLC-UM PDU is received from the lower layer
- **THEN** the system SHALL remove the 1-byte header and deliver the SDU to the upper layer
