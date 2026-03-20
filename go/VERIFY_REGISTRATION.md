# Verifying Happy-Path Registration And Initial PDU Session

This guide explains how to verify the supported Go happy path using `nr-gnb` and `nr-ue` against the live network functions already running on the machine.

Current verified boundary:
- UE registration completes successfully against live free5GC containers.
- The UE sends `Registration Complete`.
- The UE sends `PDU Session Establishment Request`.
- The UE receives `PDU Session Establishment Accept`.
- Full user-plane readiness is not yet complete, so this guide stops at session-accept signaling rather than IP traffic verification.

## Prerequisites

1.  **Live NFs**: Ensure AMF, SMF, UPF, and the rest of the required core network functions are already running on the machine.
2.  **Docker visibility**: Confirm the NF containers are present before testing.
3.  **Go Environment**: Go 1.21+ installed.
4.  **Network**: The host must have connectivity to the NF containers (default AMF IP is expected at `10.100.200.16`).

Check the NF containers first:

```bash
docker ps
```

The expected baseline is that the AMF, SMF, and UPF containers are already running before the Go gNB or UE is started.

## 1. Build the Components

From the project root, build both the gNodeB and the UE:

```bash
cd go
GOCACHE=/tmp/ueransim-go-cache GOMODCACHE=/tmp/ueransim-go-modcache go build -o /tmp/nr-gnb ./cmd/nr-gnb
GOCACHE=/tmp/ueransim-go-cache GOMODCACHE=/tmp/ueransim-go-modcache go build -o /tmp/nr-ue ./cmd/nr-ue
```

## 2. Start the gNodeB (gNB)

Run the gNodeB first. It will establish an SCTP connection with the AMF and start the Radio Link Simulation (RLS) service to listen for UEs.

```bash
/tmp/nr-gnb -config /root/proj/go/ueransim-go/config/free5gc-gnb-go.yaml
```

**Successful Startup Indicators:**
- `SCTP connected`: Confirms the N2 interface to the AMF is UP.
- `NGAP task started, sending NGSetupRequest`: The gNB is identifying itself to the core.
- `received NGAP message from AMF`: Confirms the AMF accepted the gNB (`NGSetupResponse`).

## 3. Start the User Equipment (UE)

In a separate terminal, start the UE. This will trigger a NAS Registration Request over the simulated radio interface (RLS).

```bash
/tmp/nr-ue -config /root/proj/go/ueransim-go/config/free5gc-ue-go.yaml
```

**UE Activity Indicators:**
- `sending Registration Request`: NAS layer is initiating the 5GMM procedure.
- `received NAS PDU, wrapping in RRC`: RRC layer is preparing the message for transport.

## 4. Observe The Flow

### gNodeB Logs
The gNB will log the relaying of the message from the UE to the AMF:
- `received radio packet from UE`: The RLS layer received the UDP packet.
- `decoded RLS message`: The RLS header was parsed successfully.
- `received NAS PDU from RLS, sending InitialUEMessage`: The gNB extracted the first NAS PDU and is forwarding it via NGAP.
- `received NAS PDU from RLS, sending UplinkNASTransport`: The gNB is forwarding follow-on protected NAS messages without additional parity-only handling.

### UE Logs
The UE should show the supported handwritten state progression:
- `sending Registration Request`
- `handling Authentication Request`
- `sending Authentication Response`
- `handling Security Mode Command`
- `sending Security Mode Complete`
- `handling Identity Request`
- `sending Identity Response`
- `Registration Accept received`
- `sending Registration Complete`
- `sending PDU Session Establishment Request`
- `PDU Session Establishment Accept received`
- It is also expected to receive `Configuration Update Command` from AMF in this phase. The current Go UE does not yet complete the full follow-up behavior for that command.

### 5G Core Logs
Check the AMF logs to see the core network's perspective:

```bash
docker logs amf --tail 20
```

Check SMF or related session-management logs as needed:

```bash
docker logs smf --tail 20
```

**Verification Milestones**
1.  **NF availability**: `docker ps` shows the expected core-network containers before the test starts.
2.  **Identity Match**: `MobileIdentity5GS: SUCI[...]` confirms the UE identity encoding is accepted by the core.
3.  **Authentication Start**: AMF logs show the UE transitions into authentication.
4.  **Security Activation**: UE logs show Security Mode Command handling and Security Mode Complete transmission.
5.  **Registration Completion**: UE logs show `Registration Accept received` followed by `sending Registration Complete`.
6.  **PDU Session Trigger**: UE logs show `sending PDU Session Establishment Request`.
7.  **PDU Session Result**: UE logs show `PDU Session Establishment Accept received`.
8.  **Expected current limit**: No user-plane traffic check is expected to pass yet, because the bearer/context setup work is not finished.

## 5. Trigger Initial Registration Manually

This is the shortest repeatable sequence to trigger a fresh initial registration yourself.

1. Confirm the core NFs are already up:

```bash
docker ps
```

2. Build fresh binaries:

```bash
cd /root/proj/go/ueransim-go/go
GOCACHE=/tmp/ueransim-go-cache GOMODCACHE=/tmp/ueransim-go-modcache go build -o /tmp/nr-gnb ./cmd/nr-gnb
GOCACHE=/tmp/ueransim-go-cache GOMODCACHE=/tmp/ueransim-go-modcache go build -o /tmp/nr-ue ./cmd/nr-ue
```

3. Stop any leftover Go gNB or Go UE processes from a previous run:

```bash
pkill -f '/tmp/nr-gnb -config /root/proj/go/ueransim-go/config/free5gc-gnb-go.yaml' || true
pkill -f '/tmp/nr-ue -config /root/proj/go/ueransim-go/config/free5gc-ue-go.yaml' || true
```

4. Start the gNB in one terminal:

```bash
/tmp/nr-gnb -config /root/proj/go/ueransim-go/config/free5gc-gnb-go.yaml
```

5. Start the UE in another terminal:

```bash
/tmp/nr-ue -config /root/proj/go/ueransim-go/config/free5gc-ue-go.yaml
```

That UE start is what triggers a fresh initial registration attempt.

6. Watch the core logs if you want to confirm progression:

```bash
docker logs amf --tail 60
docker logs smf --tail 60
```

## Troubleshooting

- **NF containers missing**: Run `docker ps` and ensure AMF, SMF, and UPF are up before starting the Go components.
- **No Connection to AMF**: Check that the `amfConfigs.address` in `config/free5gc-gnb-go.yaml` matches your AMF's IP.
- **UE Packets Not Received**: Ensure `linkIp` in `config/free5gc-gnb-go.yaml` is set to `0.0.0.0` to listen on all local interfaces.
- **Registration Rejection**: Check AMF logs for `UESecurityCapability is nil`, identity errors, or NAS security failures.
- **Identity failure**: If AMF rejects PEI/IMEI, check the configured IMEI value and its checksum handling in `config/free5gc-ue-go.yaml`.
- **No PDU session result**: Check SMF logs and UE logs to see whether the UL NAS Transport was accepted and whether a DL NAS Transport carrying the session response was returned.
- **No user-plane traffic**: That is currently expected. Session signaling reaches accept, but bearer/context completion is still pending.
