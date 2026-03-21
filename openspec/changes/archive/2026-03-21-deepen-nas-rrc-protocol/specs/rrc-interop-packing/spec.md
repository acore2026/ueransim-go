## ADDED Requirements

### Requirement: Authentic RRC Bit-Packing
The RRC layer SHALL construct control plane messages using the 3GPP TS 38.331 Unaligned Packed Encoding Rules (UPER) format for interoperability with external gNodeB and Core Network functions.

#### Scenario: RRCSetupComplete bit-stream generation
- **WHEN** constructing an `RRCSetupComplete` message with a NAS PDU
- **THEN** the system SHALL produce a bit-stream that matches the exact structure of `UL-DCCH-Message -> c1 -> rrcSetupComplete` as defined in 3GPP TS 38.331

### Requirement: Removal of Non-Standard RRC
The system SHALL NOT use the legacy "SimpleRRC" TLV wrapper for RRC message transport.

#### Scenario: RRC message relaying to gNodeB
- **WHEN** the NAS layer sends a PDU to the RRC layer for transport
- **THEN** the RRC layer SHALL encapsulate the PDU directly into an authentic bit-packed RRC message and relay it to the Radio Link Simulation (RLS) interface
