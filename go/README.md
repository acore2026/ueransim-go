# Go Rewrite Scaffold

This subtree contains the initial Go runtime scaffold for migrating UERANSIM from C++ to Go.

Current scope:

- UE and gNB bootstrap binaries
- shared runtime/task model
- YAML config loading for existing sample configs
- structured logging

Not implemented yet:

- SCTP, NGAP, RRC, NAS, GTP-U, TUN, radio simulation
- protocol codecs
- feature parity with the C++ implementation
