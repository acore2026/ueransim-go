package gnbctx

import "testing"

func TestSessionStoreUpsertAndLookup(t *testing.T) {
	store := NewSessionStore()
	session := store.Upsert(SessionSetupRequest{
		RANUENGAPID: 11,
		AMFUENGAPID: 22,
		SessionID:   1,
		RemoteIP:    "10.0.0.8",
		RemoteTEID:  0x12345678,
		QFIs:        []uint8{9},
	}, "10.100.200.1")

	if session.LocalTEID == 0 {
		t.Fatal("expected allocated local TEID")
	}
	bySession, ok := store.BySessionID(1)
	if !ok {
		t.Fatal("session lookup failed")
	}
	if bySession.RemoteIP != "10.0.0.8" || bySession.RemoteTEID != 0x12345678 {
		t.Fatalf("unexpected session data: %+v", bySession)
	}
	byTEID, ok := store.ByLocalTEID(session.LocalTEID)
	if !ok {
		t.Fatal("local TEID lookup failed")
	}
	if byTEID.SessionID != 1 {
		t.Fatalf("unexpected session ID: %+v", byTEID)
	}
}
