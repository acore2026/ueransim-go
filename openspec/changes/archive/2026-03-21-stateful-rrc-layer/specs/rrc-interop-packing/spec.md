## MODIFIED Requirements

### Requirement: Authentic RRC Bit-Packing
The RRC layer SHALL construct control plane messages using the 3GPP TS 38.331 Unaligned Packed Encoding Rules (UPER) format for interoperability with external gNodeB and Core Network functions.

#### Scenario: RRCSetupComplete bit-stream generation
- **WHEN** constructing an `RRCSetupComplete` message with a NAS PDU
- **THEN** the system SHALL produce a bit-stream that matches the exact structure of `UL-DCCH-Message -> c1 -> rrcSetupComplete` as defined in 3GPP TS 38.331

#### Scenario: RRCReconfigurationComplete bit-stream generation
- **WHEN** constructing an `RRCReconfigurationComplete` message
- **THEN** the system SHALL produce a bit-stream that matches the exact structure of `UL-DCCH-Message -> c1 -> rrcReconfigurationComplete` as defined in 3GPP TS 38.331

## ADDED Requirements

### Requirement: Authentic RRC Bit-Extraction
The RRC layer SHALL extract Information Elements from downlink control plane messages using bit-accurate offsets defined in 3GPP TS 38.331.

#### Scenario: RRCSetup extraction
- **WHEN** receiving a DL-CCCH bit-stream containing an `RRCSetup` message
- **THEN** the system SHALL correctly identify the message type and extract relevant configuration parameters
