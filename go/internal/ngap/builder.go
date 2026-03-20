package ngap

import (
	"encoding/binary"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap"
	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
)

type SessionResourceSetup struct {
	PDUSessionID uint8
	NASPDU       []byte
	RemoteGTPIP  string
	RemoteTEID   uint32
	QFIs         []uint8
}

type InitialContextSetupData struct {
	AMFUENGAPID int64
	RANUENGAPID int64
	NASPDU      []byte
	Sessions    []SessionResourceSetup
}

func BuildNGSetupRequest(gnbName string, gnbID []byte, bitLength uint64, plmnID []byte) (*ngapType.NGAPPDU, error) {
	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeNGSetup},
			Criticality:   ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentNGSetupRequest,
				NGSetupRequest: &ngapType.NGSetupRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerNGSetupRequestIEs{
						List: []ngapType.NGSetupRequestIEs{
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDGlobalRANNodeID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.NGSetupRequestIEsValue{
									Present: ngapType.NGSetupRequestIEsPresentGlobalRANNodeID,
									GlobalRANNodeID: &ngapType.GlobalRANNodeID{
										Present: ngapType.GlobalRANNodeIDPresentGlobalGNBID,
										GlobalGNBID: &ngapType.GlobalGNBID{
											PLMNIdentity: ngapType.PLMNIdentity{Value: aper.OctetString(plmnID)},
											GNBID: ngapType.GNBID{
												Present: ngapType.GNBIDPresentGNBID,
												GNBID: &aper.BitString{
													Bytes:     gnbID,
													BitLength: bitLength,
												},
											},
										},
									},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDSupportedTAList},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.NGSetupRequestIEsValue{
									Present: ngapType.NGSetupRequestIEsPresentSupportedTAList,
									SupportedTAList: &ngapType.SupportedTAList{
										List: []ngapType.SupportedTAItem{
											{
												TAC: ngapType.TAC{Value: aper.OctetString([]byte{0x00, 0x00, 0x01})},
												BroadcastPLMNList: ngapType.BroadcastPLMNList{
													List: []ngapType.BroadcastPLMNItem{
														{
															PLMNIdentity: ngapType.PLMNIdentity{Value: aper.OctetString(plmnID)},
															TAISliceSupportList: ngapType.SliceSupportList{
																List: []ngapType.SliceSupportItem{
																	{
																		SNSSAI: ngapType.SNSSAI{
																			SST: ngapType.SST{Value: []byte{0x01}},
																			SD:  &ngapType.SD{Value: []byte{0x01, 0x02, 0x03}},
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
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANNodeName},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
								Value: ngapType.NGSetupRequestIEsValue{
									Present:     ngapType.NGSetupRequestIEsPresentRANNodeName,
									RANNodeName: &ngapType.RANNodeName{Value: gnbName},
								},
							},
						},
					},
				},
			},
		},
	}
	return pdu, nil
}

func BuildUserLocationInformationNR(plmnID []byte, tac []byte, nrCellID []byte) *ngapType.UserLocationInformation {
	return &ngapType.UserLocationInformation{
		Present: ngapType.UserLocationInformationPresentUserLocationInformationNR,
		UserLocationInformationNR: &ngapType.UserLocationInformationNR{
			NRCGI: ngapType.NRCGI{
				PLMNIdentity: ngapType.PLMNIdentity{Value: aper.OctetString(plmnID)},
				NRCellIdentity: ngapType.NRCellIdentity{
					Value: aper.BitString{
						Bytes:     nrCellID,
						BitLength: 36,
					},
				},
			},
			TAI: ngapType.TAI{
				PLMNIdentity: ngapType.PLMNIdentity{Value: aper.OctetString(plmnID)},
				TAC:          ngapType.TAC{Value: aper.OctetString(tac)},
			},
		},
	}
}

func BuildInitialUEMessage(ranUeNgapID int64, nasPdu []byte, userLocationRefer *ngapType.UserLocationInformation) (*ngapType.NGAPPDU, error) {
	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeInitialUEMessage},
			Criticality:   ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentInitialUEMessage,
				InitialUEMessage: &ngapType.InitialUEMessage{
					ProtocolIEs: ngapType.ProtocolIEContainerInitialUEMessageIEs{
						List: []ngapType.InitialUEMessageIEs{
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.InitialUEMessageIEsValue{
									Present:     ngapType.InitialUEMessageIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ranUeNgapID},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDNASPDU},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.InitialUEMessageIEsValue{
									Present: ngapType.InitialUEMessageIEsPresentNASPDU,
									NASPDU:  &ngapType.NASPDU{Value: aper.OctetString(nasPdu)},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDUserLocationInformation},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.InitialUEMessageIEsValue{
									Present:                 ngapType.InitialUEMessageIEsPresentUserLocationInformation,
									UserLocationInformation: userLocationRefer,
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRRCEstablishmentCause},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
								Value: ngapType.InitialUEMessageIEsValue{
									Present:               ngapType.InitialUEMessageIEsPresentRRCEstablishmentCause,
									RRCEstablishmentCause: &ngapType.RRCEstablishmentCause{Value: ngapType.RRCEstablishmentCausePresentMoSignalling},
								},
							},
						},
					},
				},
			},
		},
	}
	return pdu, nil
}

func BuildUplinkNASTransport(ranUeNgapID int64, amfUeNgapID int64, nasPdu []byte, userLocation *ngapType.UserLocationInformation) (*ngapType.NGAPPDU, error) {
	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeUplinkNASTransport},
			Criticality:   ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUplinkNASTransport,
				UplinkNASTransport: &ngapType.UplinkNASTransport{
					ProtocolIEs: ngapType.ProtocolIEContainerUplinkNASTransportIEs{
						List: []ngapType.UplinkNASTransportIEs{
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.UplinkNASTransportIEsValue{
									Present:     ngapType.UplinkNASTransportIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: amfUeNgapID},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.UplinkNASTransportIEsValue{
									Present:     ngapType.UplinkNASTransportIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ranUeNgapID},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDNASPDU},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
								Value: ngapType.UplinkNASTransportIEsValue{
									Present: ngapType.UplinkNASTransportIEsPresentNASPDU,
									NASPDU:  &ngapType.NASPDU{Value: aper.OctetString(nasPdu)},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDUserLocationInformation},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
								Value: ngapType.UplinkNASTransportIEsValue{
									Present:                 ngapType.UplinkNASTransportIEsPresentUserLocationInformation,
									UserLocationInformation: userLocation,
								},
							},
						},
					},
				},
			},
		},
	}
	return pdu, nil
}

func Encode(pdu *ngapType.NGAPPDU) ([]byte, error) {
	return ngap.Encoder(*pdu)
}

func Decode(data []byte) (*ngapType.NGAPPDU, error) {
	pdu, err := ngap.Decoder(data)
	if err != nil {
		return nil, err
	}
	return pdu, nil
}

func GetNasPdu(pdu *ngapType.NGAPPDU) []byte {
	if pdu.Present != ngapType.NGAPPDUPresentInitiatingMessage {
		return nil
	}
	ini := pdu.InitiatingMessage
	switch ini.ProcedureCode.Value {
	case ngapType.ProcedureCodeDownlinkNASTransport:
		down := ini.Value.DownlinkNASTransport
		for _, ie := range down.ProtocolIEs.List {
			if ie.Id.Value == ngapType.ProtocolIEIDNASPDU {
				return []byte(ie.Value.NASPDU.Value)
			}
		}
	case ngapType.ProcedureCodeInitialContextSetup:
		req := ini.Value.InitialContextSetupRequest
		for _, ie := range req.ProtocolIEs.List {
			switch ie.Id.Value {
			case ngapType.ProtocolIEIDNASPDU:
				return []byte(ie.Value.NASPDU.Value)
			case ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtReq:
				for _, item := range ie.Value.PDUSessionResourceSetupListCxtReq.List {
					if item.NASPDU != nil {
						return []byte(item.NASPDU.Value)
					}
				}
			}
		}
	}
	return nil
}

func ParseInitialContextSetupRequest(pdu *ngapType.NGAPPDU) (*InitialContextSetupData, error) {
	if pdu.Present != ngapType.NGAPPDUPresentInitiatingMessage {
		return nil, nil
	}
	ini := pdu.InitiatingMessage
	if ini.ProcedureCode.Value != ngapType.ProcedureCodeInitialContextSetup {
		return nil, nil
	}

	req := ini.Value.InitialContextSetupRequest
	result := &InitialContextSetupData{}
	for _, ie := range req.ProtocolIEs.List {
		switch ie.Id.Value {
		case ngapType.ProtocolIEIDAMFUENGAPID:
			result.AMFUENGAPID = ie.Value.AMFUENGAPID.Value
		case ngapType.ProtocolIEIDRANUENGAPID:
			result.RANUENGAPID = ie.Value.RANUENGAPID.Value
		case ngapType.ProtocolIEIDNASPDU:
			result.NASPDU = append([]byte(nil), ie.Value.NASPDU.Value...)
		case ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtReq:
			for _, item := range ie.Value.PDUSessionResourceSetupListCxtReq.List {
				transfer := ngapType.PDUSessionResourceSetupRequestTransfer{}
				raw := append([]byte(nil), item.PDUSessionResourceSetupRequestTransfer...)
				if err := aper.UnmarshalWithParams(raw, &transfer, "valueExt"); err != nil {
					return nil, err
				}

				setup := SessionResourceSetup{
					PDUSessionID: uint8(item.PDUSessionID.Value),
				}
				if item.NASPDU != nil {
					setup.NASPDU = append([]byte(nil), item.NASPDU.Value...)
				}

				for _, transferIE := range transfer.ProtocolIEs.List {
					switch transferIE.Id.Value {
					case ngapType.ProtocolIEIDULNGUUPTNLInformation:
						info := transferIE.Value.ULNGUUPTNLInformation
						if info != nil && info.Present == ngapType.UPTransportLayerInformationPresentGTPTunnel {
							ipv4, _ := ngapConvert.IPAddressToString(info.GTPTunnel.TransportLayerAddress)
							setup.RemoteGTPIP = ipv4
							setup.RemoteTEID = binary.BigEndian.Uint32(info.GTPTunnel.GTPTEID.Value)
						}
					case ngapType.ProtocolIEIDQosFlowSetupRequestList:
						if transferIE.Value.QosFlowSetupRequestList != nil {
							for _, item := range transferIE.Value.QosFlowSetupRequestList.List {
								setup.QFIs = append(setup.QFIs, uint8(item.QosFlowIdentifier.Value))
							}
						}
					}
				}

				if result.NASPDU == nil && len(setup.NASPDU) > 0 {
					result.NASPDU = append([]byte(nil), setup.NASPDU...)
				}
				result.Sessions = append(result.Sessions, setup)
			}
		}
	}

	return result, nil
}

type SessionResourceSetupResponse struct {
	PDUSessionID uint8
	LocalGTPIP   string
	LocalTEID    uint32
	QFIs         []uint8
}

func BuildInitialContextSetupResponse(amfUeNgapID, ranUeNgapID int64, sessions []SessionResourceSetupResponse) (*ngapType.NGAPPDU, error) {
	pdu := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentSuccessfulOutcome,
		SuccessfulOutcome: &ngapType.SuccessfulOutcome{
			ProcedureCode: ngapType.ProcedureCode{Value: ngapType.ProcedureCodeInitialContextSetup},
			Criticality:   ngapType.Criticality{Value: ngapType.CriticalityPresentReject},
			Value: ngapType.SuccessfulOutcomeValue{
				Present: ngapType.SuccessfulOutcomePresentInitialContextSetupResponse,
				InitialContextSetupResponse: &ngapType.InitialContextSetupResponse{
					ProtocolIEs: ngapType.ProtocolIEContainerInitialContextSetupResponseIEs{
						List: []ngapType.InitialContextSetupResponseIEs{
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
								Value: ngapType.InitialContextSetupResponseIEsValue{
									Present:     ngapType.InitialContextSetupResponseIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: amfUeNgapID},
								},
							},
							{
								Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
								Value: ngapType.InitialContextSetupResponseIEsValue{
									Present:     ngapType.InitialContextSetupResponseIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ranUeNgapID},
								},
							},
						},
					},
				},
			},
		},
	}

	if len(sessions) > 0 {
		setupList := &ngapType.PDUSessionResourceSetupListCxtRes{}
		for _, session := range sessions {
			transfer, err := buildPDUSessionResourceSetupResponseTransfer(session)
			if err != nil {
				return nil, err
			}
			setupList.List = append(setupList.List, ngapType.PDUSessionResourceSetupItemCxtRes{
				PDUSessionID:                            ngapType.PDUSessionID{Value: int64(session.PDUSessionID)},
				PDUSessionResourceSetupResponseTransfer: transfer,
			})
		}
		pdu.SuccessfulOutcome.Value.InitialContextSetupResponse.ProtocolIEs.List = append(
			pdu.SuccessfulOutcome.Value.InitialContextSetupResponse.ProtocolIEs.List,
			ngapType.InitialContextSetupResponseIEs{
				Id:          ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtRes},
				Criticality: ngapType.Criticality{Value: ngapType.CriticalityPresentIgnore},
				Value: ngapType.InitialContextSetupResponseIEsValue{
					Present:                           ngapType.InitialContextSetupResponseIEsPresentPDUSessionResourceSetupListCxtRes,
					PDUSessionResourceSetupListCxtRes: setupList,
				},
			},
		)
	}

	return pdu, nil
}

func buildPDUSessionResourceSetupResponseTransfer(session SessionResourceSetupResponse) ([]byte, error) {
	associatedQFIs := ngapType.AssociatedQosFlowList{}
	for _, qfi := range session.QFIs {
		associatedQFIs.List = append(associatedQFIs.List, ngapType.AssociatedQosFlowItem{
			QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: int64(qfi)},
		})
	}
	if len(associatedQFIs.List) == 0 {
		associatedQFIs.List = append(associatedQFIs.List, ngapType.AssociatedQosFlowItem{
			QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: 1},
		})
	}

	teid := make([]byte, 4)
	binary.BigEndian.PutUint32(teid, session.LocalTEID)
	transfer := ngapType.PDUSessionResourceSetupResponseTransfer{
		DLQosFlowPerTNLInformation: ngapType.QosFlowPerTNLInformation{
			UPTransportLayerInformation: ngapType.UPTransportLayerInformation{
				Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
				GTPTunnel: &ngapType.GTPTunnel{
					TransportLayerAddress: ngapConvert.IPAddressToNgap(session.LocalGTPIP, ""),
					GTPTEID:               ngapType.GTPTEID{Value: teid},
				},
			},
			AssociatedQosFlowList: associatedQFIs,
		},
	}
	return aper.MarshalWithParams(transfer, "valueExt")
}
