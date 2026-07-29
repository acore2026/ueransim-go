package ngap

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
)

func TestEncodeNGSetupRequest(t *testing.T) {
	gnbName := "ueransim-gnb"
	gnbID := []byte{0x00, 0x01, 0x02}
	plmnID := []byte{0x02, 0xf8, 0x39}

	pdu, err := BuildNGSetupRequest(gnbName, gnbID, 24, plmnID)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	res, err := Encode(pdu)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded hex: %s", hex.EncodeToString(res))

	if len(res) == 0 {
		t.Error("encoded result is empty")
	}
}

func TestInitialUEMessage(t *testing.T) {
	plmnID := []byte{0x02, 0xf8, 0x39}
	tac := []byte{0x00, 0x00, 0x01}
	nrCellID := []byte{0x00, 0x00, 0x00, 0x00, 0x10}
	nasPdu := []byte{0x7e, 0x00, 0x41, 0x79, 0x00, 0x0d, 0x01, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x00, 0x00, 0x01, 0x23, 0x45, 0x67, 0x89}

	userLocation := BuildUserLocationInformationNR(plmnID, tac, nrCellID)
	pdu, err := BuildInitialUEMessage(1, nasPdu, userLocation)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	res, err := Encode(pdu)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded hex: %s", hex.EncodeToString(res))

	if len(res) == 0 {
		t.Error("encoded result is empty")
	}
}

func TestParseInitialContextSetupRequest(t *testing.T) {
	transfer := ngapType.PDUSessionResourceSetupRequestTransfer{
		ProtocolIEs: ngapType.ProtocolIEContainerPDUSessionResourceSetupRequestTransferIEs{
			List: []ngapType.PDUSessionResourceSetupRequestTransferIEs{
				{
					Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDULNGUUPTNLInformation},
					Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
					Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
						Present: ngapType.PDUSessionResourceSetupRequestTransferIEsPresentULNGUUPTNLInformation,
						ULNGUUPTNLInformation: &ngapType.UPTransportLayerInformation{
							Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
							GTPTunnel: &ngapType.GTPTunnel{
								TransportLayerAddress: ngapConvert.IPAddressToNgap("10.0.0.8", ""),
								GTPTEID:               ngapType.GTPTEID{Value: []byte{0x12, 0x34, 0x56, 0x78}},
							},
						},
					},
				},
			},
		},
	}
	rawTransfer, err := aper.MarshalWithParams(transfer, "valueExt")
	if err != nil {
		t.Fatalf("marshal transfer: %v", err)
	}

	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeInitialContextSetup},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentInitialContextSetupRequest,
				InitialContextSetupRequest: &ngapType.InitialContextSetupRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerInitialContextSetupRequestIEs{
						List: []ngapType.InitialContextSetupRequestIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present:     ngapType.InitialContextSetupRequestIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 1001},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present:     ngapType.InitialContextSetupRequestIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: 2002},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtReq},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present: ngapType.InitialContextSetupRequestIEsPresentPDUSessionResourceSetupListCxtReq,
									PDUSessionResourceSetupListCxtReq: &ngapType.PDUSessionResourceSetupListCxtReq{
										List: []ngapType.PDUSessionResourceSetupItemCxtReq{
											{
												PDUSessionID:                           ngapType.PDUSessionID{Value: 1},
												NASPDU:                                 &ngapType.NASPDU{Value: []byte{0x01, 0x02, 0x03}},
												PDUSessionResourceSetupRequestTransfer: rawTransfer,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	parsed, err := ParseInitialContextSetupRequest(pdu)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.AMFUENGAPID != 1001 || parsed.RANUENGAPID != 2002 {
		t.Fatalf("unexpected ids: %+v", parsed)
	}
	if len(parsed.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(parsed.Sessions))
	}
	session := parsed.Sessions[0]
	if session.RemoteGTPIP != "10.0.0.8" || session.RemoteTEID != 0x12345678 {
		t.Fatalf("unexpected remote tunnel: %+v", session)
	}
	if len(session.QFIs) != 0 {
		t.Fatalf("unexpected QFIs: %+v", session.QFIs)
	}
}

func TestBuildInitialContextSetupResponse(t *testing.T) {
	pdu, err := BuildInitialContextSetupResponse(11, 22, []SessionResourceSetupResponse{
		{
			PDUSessionID: 1,
			LocalGTPIP:   "10.100.200.1",
			LocalTEID:    0x01020304,
			QFIs:         []uint8{9},
		},
	})
	if err != nil {
		t.Fatalf("build response: %v", err)
	}
	encoded, err := Encode(pdu)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Present != ngapType.NGAPPDUPresentSuccessfulOutcome {
		t.Fatalf("unexpected present: %v", decoded.Present)
	}
	resp := decoded.SuccessfulOutcome.Value.InitialContextSetupResponse
	if resp == nil {
		t.Fatal("missing InitialContextSetupResponse")
	}
	var found bool
	for _, ie := range resp.ProtocolIEs.List {
		if ie.Id.Value != ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtRes {
			continue
		}
		found = true
		item := ie.Value.PDUSessionResourceSetupListCxtRes.List[0]
		transfer := ngapType.PDUSessionResourceSetupResponseTransfer{}
		raw := append([]byte(nil), item.PDUSessionResourceSetupResponseTransfer...)
		if err := aper.UnmarshalWithParams(raw, &transfer, "valueExt"); err != nil {
			t.Fatalf("unmarshal response transfer: %v", err)
		}
		gtpTunnel := transfer.DLQosFlowPerTNLInformation.UPTransportLayerInformation.GTPTunnel
		if binary.BigEndian.Uint32(gtpTunnel.GTPTEID.Value) != 0x01020304 {
			t.Fatalf("unexpected local TEID: %x", gtpTunnel.GTPTEID.Value)
		}
	}
	if !found {
		t.Fatal("missing PDUSessionResourceSetupListCxtRes IE")
	}
}

func TestParsePDUSessionResourceSetupRequest(t *testing.T) {
	rawTransfer := buildTestPDUSessionResourceSetupRequestTransfer(t, "10.0.0.9", 0x11223344, 9)
	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodePDUSessionResourceSetup},
			Criticality:   ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentPDUSessionResourceSetupRequest,
				PDUSessionResourceSetupRequest: &ngapType.PDUSessionResourceSetupRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerPDUSessionResourceSetupRequestIEs{
						List: []ngapType.PDUSessionResourceSetupRequestIEs{
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present:     ngapType.PDUSessionResourceSetupRequestIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 1001},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present:     ngapType.PDUSessionResourceSetupRequestIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: 2002},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present: ngapType.PDUSessionResourceSetupRequestIEsPresentPDUSessionResourceSetupListSUReq,
									PDUSessionResourceSetupListSUReq: &ngapType.PDUSessionResourceSetupListSUReq{
										List: []ngapType.PDUSessionResourceSetupItemSUReq{
											{
												PDUSessionID:     ngapType.PDUSessionID{Value: 1},
												PDUSessionNASPDU: &ngapType.NASPDU{Value: []byte{0x01, 0x02, 0x03}},
												SNSSAI: ngapType.SNSSAI{
													SST: ngapType.SST{Value: []byte{0x01}},
												},
												PDUSessionResourceSetupRequestTransfer: rawTransfer,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	encoded, err := Encode(pdu)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode request: %v", err)
	}
	parsed, err := ParsePDUSessionResourceSetupRequest(decoded)
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected standalone setup request")
	}
	if parsed.AMFUENGAPID != 1001 || parsed.RANUENGAPID != 2002 {
		t.Fatalf("unexpected ids: %+v", parsed)
	}
	if len(parsed.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(parsed.Sessions))
	}
	session := parsed.Sessions[0]
	if session.PDUSessionID != 1 || session.RemoteGTPIP != "10.0.0.9" || session.RemoteTEID != 0x11223344 {
		t.Fatalf("unexpected session: %+v", session)
	}
	if len(session.QFIs) != 1 || session.QFIs[0] != 9 {
		t.Fatalf("unexpected QFIs: %+v", session.QFIs)
	}
	if got := GetNasPdu(decoded); string(got) != string([]byte{0x01, 0x02, 0x03}) {
		t.Fatalf("unexpected NAS PDU: %x", got)
	}
}

func TestBuildPDUSessionResourceSetupResponse(t *testing.T) {
	pdu, err := BuildPDUSessionResourceSetupResponse(11, 22, []SessionResourceSetupResponse{
		{
			PDUSessionID: 1,
			LocalGTPIP:   "10.100.200.1",
			LocalTEID:    0x01020304,
			QFIs:         []uint8{9},
		},
	})
	if err != nil {
		t.Fatalf("build response: %v", err)
	}
	encoded, err := Encode(pdu)
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Present != ngapType.NGAPPDUPresentSuccessfulOutcome {
		t.Fatalf("unexpected present: %v", decoded.Present)
	}
	if decoded.SuccessfulOutcome.ProcedureCode.Value != ngapType.ProcedureCodePDUSessionResourceSetup {
		t.Fatalf("unexpected procedure: %d", decoded.SuccessfulOutcome.ProcedureCode.Value)
	}
	resp := decoded.SuccessfulOutcome.Value.PDUSessionResourceSetupResponse
	if resp == nil {
		t.Fatal("missing PDUSessionResourceSetupResponse")
	}
	for _, ie := range resp.ProtocolIEs.List {
		if ie.Id.Value != ngapType.ProtocolIEIDPDUSessionResourceSetupListSURes {
			continue
		}
		if len(ie.Value.PDUSessionResourceSetupListSURes.List) != 1 {
			t.Fatalf("unexpected setup response list: %+v", ie.Value.PDUSessionResourceSetupListSURes.List)
		}
		item := ie.Value.PDUSessionResourceSetupListSURes.List[0]
		transfer := ngapType.PDUSessionResourceSetupResponseTransfer{}
		if err := aper.UnmarshalWithParams(
			append([]byte(nil), item.PDUSessionResourceSetupResponseTransfer...),
			&transfer,
			"valueExt",
		); err != nil {
			t.Fatalf("unmarshal response transfer: %v", err)
		}
		tunnel := transfer.DLQosFlowPerTNLInformation.UPTransportLayerInformation.GTPTunnel
		if tunnel == nil || binary.BigEndian.Uint32(tunnel.GTPTEID.Value) != 0x01020304 {
			t.Fatalf("unexpected response tunnel: %+v", tunnel)
		}
		return
	}
	t.Fatal("missing PDUSessionResourceSetupListSURes IE")
}

func buildTestPDUSessionResourceSetupRequestTransfer(
	t *testing.T,
	remoteIP string,
	remoteTEID uint32,
	qfi int64,
) []byte {
	t.Helper()
	teid := make([]byte, 4)
	binary.BigEndian.PutUint32(teid, remoteTEID)
	transfer := ngapType.PDUSessionResourceSetupRequestTransfer{
		ProtocolIEs: ngapType.ProtocolIEContainerPDUSessionResourceSetupRequestTransferIEs{
			List: []ngapType.PDUSessionResourceSetupRequestTransferIEs{
				{
					Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDULNGUUPTNLInformation},
					Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
					Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
						Present: ngapType.PDUSessionResourceSetupRequestTransferIEsPresentULNGUUPTNLInformation,
						ULNGUUPTNLInformation: &ngapType.UPTransportLayerInformation{
							Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
							GTPTunnel: &ngapType.GTPTunnel{
								TransportLayerAddress: ngapConvert.IPAddressToNgap(remoteIP, ""),
								GTPTEID:               ngapType.GTPTEID{Value: teid},
							},
						},
					},
				},
				{
					Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDQosFlowSetupRequestList},
					Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
					Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
						Present: ngapType.PDUSessionResourceSetupRequestTransferIEsPresentQosFlowSetupRequestList,
						QosFlowSetupRequestList: &ngapType.QosFlowSetupRequestList{
							List: []ngapType.QosFlowSetupRequestItem{
								{
									QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: qfi},
									QosFlowLevelQosParameters: ngapType.QosFlowLevelQosParameters{
										QosCharacteristics: ngapType.QosCharacteristics{
											Present: ngapType.QosCharacteristicsPresentNonDynamic5QI,
											NonDynamic5QI: &ngapType.NonDynamic5QIDescriptor{
												FiveQI: ngapType.FiveQI{Value: 9},
											},
										},
										AllocationAndRetentionPriority: ngapType.AllocationAndRetentionPriority{
											PriorityLevelARP: ngapType.PriorityLevelARP{Value: 8},
											PreEmptionCapability: ngapType.PreEmptionCapability{
												Value: ngapType.PreEmptionCapabilityPresentShallNotTriggerPreEmption,
											},
											PreEmptionVulnerability: ngapType.PreEmptionVulnerability{
												Value: ngapType.PreEmptionVulnerabilityPresentNotPreEmptable,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	raw, err := aper.MarshalWithParams(transfer, "valueExt")
	if err != nil {
		t.Fatalf("marshal setup request transfer: %v", err)
	}
	return raw
}
