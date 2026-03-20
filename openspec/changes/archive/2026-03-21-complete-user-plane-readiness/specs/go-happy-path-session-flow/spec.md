## MODIFIED Requirements

### Requirement: Go rewrite SHALL support a documented happy-path registration and session flow
The Go implementation SHALL support one documented end-to-end flow covering UE startup, radio attachment, initial registration request delivery, authentication challenge handling, NAS security activation, registration completion, PDU session establishment, gNB-side bearer and tunnel setup, and basic user-plane readiness against the live network functions available on the machine.

#### Scenario: Successful end-to-end happy path with usable session
- **WHEN** the Go gNB and Go UE are started with the existing compatible NF containers running on the machine
- **THEN** the system SHALL progress through the documented registration and PDU session flow and reach a basic user-plane-ready state without requiring the C++ implementation

### Requirement: The supported flow SHALL have a concrete verification path
The project SHALL document at least one repeatable verification path for the happy-path registration and PDU session flow against the live NF environment on the machine, including the expected milestones for successful progression through registration, session setup, bearer completion, and basic user-plane readiness.

#### Scenario: Verifying the supported flow through user-plane readiness
- **WHEN** a contributor follows the documented verification procedure
- **THEN** they SHALL be able to observe the expected registration, session-establishment, bearer-setup, and user-plane-readiness milestones for the supported happy path

## ADDED Requirements

### Requirement: gNB SHALL extract session resource setup details from InitialContextSetupRequest
The Go gNB SHALL parse the PDU session resource setup content carried in `InitialContextSetupRequest` sufficiently to learn the tunnel parameters required for the supported happy path, including the remote GTP-U endpoint information needed for user-plane setup.

#### Scenario: Parsing bearer setup details from InitialContextSetupRequest
- **WHEN** the AMF sends an `InitialContextSetupRequest` containing PDU session resource setup content for the supported happy path
- **THEN** the gNB SHALL extract the remote tunnel information needed to complete session setup rather than handling only the embedded NAS PDU

### Requirement: gNB SHALL maintain explicit PDU session state for the supported happy path
The Go gNB SHALL create and maintain explicit PDU session state for the supported happy path, including UE association, PDU session identity, remote tunnel parameters, locally allocated tunnel information, and the mappings needed for uplink and downlink packet routing.

#### Scenario: Creating session state after receiving session resource setup
- **WHEN** the gNB has parsed a supported PDU session resource setup request
- **THEN** it SHALL create or update a session record that can be used by both control-plane completion and GTP-U forwarding logic

### Requirement: gNB SHALL provide a minimal real GTP-U forwarding path
The Go implementation SHALL replace the placeholder gNB GTP task with a minimal real GTP-U path for the supported happy path, including uplink encapsulation from UE-originated packets and downlink decapsulation toward the matching UE session.

#### Scenario: Forwarding uplink and downlink packets through the supported session
- **WHEN** a supported PDU session has completed setup and packets are exchanged for that session
- **THEN** the gNB SHALL use the established tunnel mappings to forward uplink packets toward the UPF and deliver downlink packets back to the matching UE path

### Requirement: UE traffic SHALL be connected to the established session only after bearer setup completes
The Go implementation SHALL connect UE TUN traffic to the supported PDU session only after the required bearer and tunnel setup information has been established, so packets are not emitted on an incomplete session path.

#### Scenario: Gating UE traffic on session readiness
- **WHEN** registration and `PDU Session Establishment Accept` have completed but bearer setup is not yet complete
- **THEN** the implementation SHALL avoid treating the session as user-plane ready

### Requirement: gNB SHALL send the required NGAP bearer/setup response for the supported flow
The Go gNB SHALL send the required NGAP response for the supported session and bearer setup path so that the RAN side of the documented happy path is complete.

#### Scenario: Completing the RAN-side setup response
- **WHEN** the gNB has allocated the local tunnel state required for the supported session
- **THEN** it SHALL send the corresponding NGAP setup response for that supported happy-path session
