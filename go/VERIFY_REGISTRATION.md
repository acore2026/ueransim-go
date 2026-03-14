# Verifying Initial Registration

This guide explains how to verify the 5G registration flow using the Go implementation of `nr-gnb` and `nr-ue` against a running 5G Core (e.g., free5GC).

## Prerequisites

1.  **5G Core**: Ensure AMF, SMF, UPF, and other core network functions are running (e.g., via Docker).
2.  **Go Environment**: Go 1.21+ installed.
3.  **Network**: The host must have connectivity to the 5G Core containers (default AMF IP is expected at `10.100.200.16`).

## 1. Build the Components

From the project root, build both the gNodeB and the UE:

```bash
cd go
go build -o ../build/nr-gnb cmd/nr-gnb/main.go
go build -o ../build/nr-ue cmd/nr-ue/main.go
cd ..
```

## 2. Start the gNodeB (gNB)

Run the gNodeB first. It will establish an SCTP connection with the AMF and start the Radio Link Simulation (RLS) service to listen for UEs.

```bash
./build/nr-gnb -config config/free5gc-gnb-go.yaml
```

**Successful Startup Indicators:**
- `SCTP connected`: Confirms the N2 interface to the AMF is UP.
- `NGAP task started, sending NGSetupRequest`: The gNB is identifying itself to the core.
- `received NGAP message from AMF`: Confirms the AMF accepted the gNB (`NGSetupResponse`).

## 3. Start the User Equipment (UE)

In a separate terminal, start the UE. This will trigger a NAS Registration Request over the simulated radio interface (RLS).

```bash
./build/nr-ue -config config/free5gc-ue-go.yaml
```

**UE Activity Indicators:**
- `sending Registration Request`: NAS layer is initiating the 5GMM procedure.
- `received NAS PDU, wrapping in RRC`: RRC layer is preparing the message for transport.

## 4. Observe the Flow

### gNodeB Logs
The gNB will log the relaying of the message from the UE to the AMF:
- `received radio packet from UE`: The RLS layer received the UDP packet.
- `decoded RLS message`: The RLS header was parsed successfully.
- `received NAS PDU from RLS, sending InitialUEMessage`: The gNB extracted the NAS PDU and is now forwarding it to the AMF via NGAP.

### 5G Core (AMF) Logs
Check the AMF logs to see the core network's perspective:

```bash
docker logs amf --tail 20
```

**Verification Milestones:**
1.  **Identity Match**: `MobileIdentity5GS: SUCI[suci-0-208-93-0000-0-0-0000000010]` confirms the UE's BCD encoding is correct.
2.  **Authentication Start**: `Handle event[Start Authentication], transition from [Deregistered] to [Authentication]` confirms the core has accepted the registration request and is challenging the UE.

## Troubleshooting

- **No Connection to AMF**: Check that the `amfConfigs.address` in `config/free5gc-gnb-go.yaml` matches your AMF's IP.
- **UE Packets Not Received**: Ensure `linkIp` in `config/free5gc-gnb-go.yaml` is set to `0.0.0.0` to listen on all local interfaces.
- **Registration Rejection**: Check AMF logs for "UESecurityCapability is nil" or identity errors, which would indicate a protocol encoding issue.
