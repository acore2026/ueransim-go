package tasks

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/lib/sctp"
	localngap "github.com/acore2026/ueransim-go/internal/ngap"
	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
)

type recordingTaskHandler struct {
	received chan runtime.Message
}

func (h *recordingTaskHandler) OnStart(context.Context, *runtime.Task) error { return nil }
func (h *recordingTaskHandler) OnStop(context.Context) error                 { return nil }
func (h *recordingTaskHandler) OnMessage(_ context.Context, msg runtime.Message) error {
	h.received <- msg
	return nil
}

func TestGnbNgapHandlesPDUSessionResourceSetupRequest(t *testing.T) {
	logger := logging.New("test-gnb-ngap")
	sctpRecorder := &recordingTaskHandler{received: make(chan runtime.Message, 1)}
	pdcpRecorder := &recordingTaskHandler{received: make(chan runtime.Message, 1)}
	sctpTask := runtime.NewTask("test-sctp-recorder", logger, sctpRecorder, 4)
	pdcpTask := runtime.NewTask("test-pdcp-recorder", logger, pdcpRecorder, 4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = sctpTask.Run(ctx) }()
	go func() { _ = pdcpTask.Run(ctx) }()

	sessionStore := gnbctx.NewSessionStore()
	handler := NewGnbNgapTaskHandler(
		logger,
		"test-gnb",
		nil,
		nil,
		nil,
		"10.100.200.1",
		sctpTask,
		pdcpTask,
		sessionStore,
	)

	request := buildStandaloneSetupRequest(t)
	encodedRequest, err := localngap.Encode(request)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	if err := handler.OnMessage(ctx, runtime.Message{
		Type: sctp.MessageTypeSctpReceive,
		Payload: sctp.ReceiveMessage{
			Stream: 7,
			Ppid:   60,
			Data:   encodedRequest,
		},
	}); err != nil {
		t.Fatalf("handle request: %v", err)
	}

	select {
	case msg := <-sctpRecorder.received:
		if msg.Type != "ngap_to_sctp" {
			t.Fatalf("unexpected SCTP task message: %s", msg.Type)
		}
		sent := msg.Payload.(sctp.SendMessage)
		if sent.Stream != 7 {
			t.Fatalf("response used stream %d, want 7", sent.Stream)
		}
		response, err := localngap.Decode(sent.Data)
		if err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Present != ngapType.NGAPPDUPresentSuccessfulOutcome ||
			response.SuccessfulOutcome.ProcedureCode.Value != ngapType.ProcedureCodePDUSessionResourceSetup ||
			response.SuccessfulOutcome.Value.PDUSessionResourceSetupResponse == nil {
			t.Fatalf("unexpected response PDU: %+v", response)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for PDUSessionResourceSetupResponse")
	}

	session, ok := sessionStore.BySessionID(1)
	if !ok {
		t.Fatal("session was not stored")
	}
	if session.AMFUENGAPID != 1001 || session.RANUENGAPID != 2002 {
		t.Fatalf("unexpected session IDs: %+v", session)
	}
	if session.RemoteIP != "10.0.0.9" || session.RemoteTEID != 0x11223344 {
		t.Fatalf("unexpected remote tunnel: %+v", session)
	}
	if session.LocalIP != "10.100.200.1" || session.LocalTEID == 0 {
		t.Fatalf("unexpected local tunnel: %+v", session)
	}
	if len(session.QFIs) != 1 || session.QFIs[0] != 9 {
		t.Fatalf("unexpected QFIs: %+v", session.QFIs)
	}

	select {
	case msg := <-pdcpRecorder.received:
		if msg.Payload == nil {
			t.Fatal("embedded NAS delivery payload is nil")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for embedded NAS delivery")
	}
}

func buildStandaloneSetupRequest(t *testing.T) *ngapType.NGAPPDU {
	t.Helper()
	teid := make([]byte, 4)
	binary.BigEndian.PutUint32(teid, 0x11223344)
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
								TransportLayerAddress: ngapConvert.IPAddressToNgap("10.0.0.9", ""),
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
									QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: 9},
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
	rawTransfer, err := aper.MarshalWithParams(transfer, "valueExt")
	if err != nil {
		t.Fatalf("marshal setup transfer: %v", err)
	}

	return &ngapType.NGAPPDU{
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
}
