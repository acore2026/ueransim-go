package gnbctx

import (
	"fmt"
	"net"
	"sync"
)

const GTPUPort = 2152

type SessionSetupRequest struct {
	RANUENGAPID int64
	AMFUENGAPID int64
	SessionID   uint8
	RemoteIP    string
	RemoteTEID  uint32
	QFIs        []uint8
}

type Session struct {
	RANUENGAPID int64
	AMFUENGAPID int64
	SessionID   uint8
	RemoteIP    string
	RemoteTEID  uint32
	LocalIP     string
	LocalTEID   uint32
	QFIs        []uint8
}

func (s Session) RemoteAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.RemoteIP, GTPUPort))
}

type SessionStore struct {
	mu          sync.RWMutex
	nextLocalTE uint32
	bySessionID map[uint8]Session
	byLocalTEID map[uint32]Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		nextLocalTE: 1,
		bySessionID: make(map[uint8]Session),
		byLocalTEID: make(map[uint32]Session),
	}
}

func (s *SessionStore) Upsert(req SessionSetupRequest, localIP string) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.bySessionID[req.SessionID]
	if !ok {
		session = Session{
			SessionID: req.SessionID,
			LocalIP:   localIP,
			LocalTEID: s.nextLocalTE,
		}
		s.nextLocalTE++
	}

	session.RANUENGAPID = req.RANUENGAPID
	session.AMFUENGAPID = req.AMFUENGAPID
	session.RemoteIP = req.RemoteIP
	session.RemoteTEID = req.RemoteTEID
	session.QFIs = append([]uint8(nil), req.QFIs...)
	session.LocalIP = localIP

	s.bySessionID[req.SessionID] = session
	s.byLocalTEID[session.LocalTEID] = session
	return session
}

func (s *SessionStore) BySessionID(id uint8) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.bySessionID[id]
	return session, ok
}

func (s *SessionStore) ByLocalTEID(teid uint32) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.byLocalTEID[teid]
	return session, ok
}

type UplinkPacket struct {
	SessionID uint8
	Data      []byte
}

type DownlinkPacket struct {
	SessionID uint8
	Data      []byte
}
