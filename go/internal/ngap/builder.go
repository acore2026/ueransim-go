package ngap

import (
	"github.com/free5gc/ngap"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/aper"
)

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

func Encode(pdu *ngapType.NGAPPDU) ([]byte, error) {
	return ngap.Encoder(*pdu)
}
