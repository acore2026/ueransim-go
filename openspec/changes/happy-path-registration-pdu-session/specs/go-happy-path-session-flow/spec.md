## ADDED Requirements

### Requirement: Go rewrite SHALL support a documented happy-path registration and session flow
The Go implementation SHALL support one documented end-to-end flow covering UE startup, radio attachment, initial registration request delivery, authentication challenge handling, NAS security activation, registration completion, PDU session establishment, and user-plane readiness against the live network functions available on the machine.

#### Scenario: Successful end-to-end happy path
- **WHEN** the Go gNB and Go UE are started with the existing compatible NF containers running on the machine
- **THEN** the system SHALL progress through the documented registration and PDU session flow without requiring the C++ implementation

### Requirement: Procedure orchestration SHALL remain handwritten
The Go implementation SHALL keep procedure orchestration in handwritten code for UE and gNB state transitions, expected message sequencing, and happy-path control decisions.

#### Scenario: Implementing registration state progression
- **WHEN** an engineer adds or modifies a supported registration step
- **THEN** the procedure progression SHALL be expressed in handwritten state-machine logic rather than generated behavior code

### Requirement: Repetitive protocol representation SHALL be generated or wrapped behind generated helpers
The Go implementation SHALL use generated or schema-driven helpers for repetitive protocol representations such as message field containers, encoding helpers, and stable message builders, and those helpers SHALL be consumed through narrow handwritten wrapper APIs.

#### Scenario: Using generated protocol structures
- **WHEN** a happy-path procedure needs ASN.1 or information-element structure handling
- **THEN** the implementation SHALL isolate schema-driven details behind focused helpers instead of scattering low-level wire-format construction across procedure code

### Requirement: Unsupported C++ complexity SHALL be explicitly excluded
The Go happy-path slice SHALL explicitly exclude advanced RRC branches, rare recovery behavior, parity-only helper layers, and other C++ complexity that is not required for the documented registration and PDU session flow.

#### Scenario: Encountering legacy C++ edge-case logic
- **WHEN** existing C++ behavior is not required to complete the supported happy path
- **THEN** the Go implementation SHALL treat that behavior as out of scope for this change rather than porting it by default

### Requirement: The supported flow SHALL have a concrete verification path
The project SHALL document at least one repeatable verification path for the happy-path registration and PDU session flow against the live NF environment on the machine, including the expected milestones for successful progression.

#### Scenario: Verifying the supported flow
- **WHEN** a contributor follows the documented verification procedure
- **THEN** they SHALL be able to observe the expected registration and session-establishment milestones for the supported happy path
