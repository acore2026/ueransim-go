package sctp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ishidawataru/sctp"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

type mockTargetTaskHandler struct {
	received chan runtime.Message
}

func newMockTargetTaskHandler() *mockTargetTaskHandler {
	return &mockTargetTaskHandler{
		received: make(chan runtime.Message, 10),
	}
}

func (h *mockTargetTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error { return nil }
func (h *mockTargetTaskHandler) OnStop(ctx context.Context) error                   { return nil }
func (h *mockTargetTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	h.received <- msg
	return nil
}

func TestSctpClientTask(t *testing.T) {
	logger := logging.New("test_sctp")

	// 1. Setup a local SCTP Server to connect to
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // SCTP uses UDP-like addressing for Resolve
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
	saddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{{IP: addr.IP}},
		Port:    addr.Port,
	}

	ln, err := sctp.ListenSCTP("sctp", saddr)
	if err != nil {
		t.Skipf("SCTP not supported or failed to listen: %v", err)
	}
	defer ln.Close()

	serverPort := ln.Addr().(*sctp.SCTPAddr).Port

	// 2. Setup Client Task
	mockHandler := newMockTargetTaskHandler()
	targetTask := runtime.NewTask("target", logger, mockHandler, 10)

	clientHandler := NewClientTaskHandler("127.0.0.1", "127.0.0.1", serverPort, targetTask, logger)
	clientTask := runtime.NewTask("sctp_client", logger, clientHandler, 10)

	group := runtime.NewGroup(logger, targetTask, clientTask)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- group.Run(ctx)
	}()

	// 3. Accept connection on server side
	serverConnRaw, err := ln.Accept()
	if err != nil {
		t.Fatalf("failed to accept: %v", err)
	}
	serverConn := serverConnRaw.(*sctp.SCTPConn)
	defer serverConn.Close()

	if err := serverConn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO); err != nil {
		t.Logf("Warning: server failed to subscribe to SCTP events: %v", err)
	}

	// 4. Test Sending from Server to Client
	testData := []byte("hello sctp client")
	info := &sctp.SndRcvInfo{
		Stream: 1,
		PPID:   1234,
	}
	_, err = serverConn.SCTPWrite(testData, info)
	if err != nil {
		t.Fatalf("failed to write to client: %v", err)
	}

	// Verify receipt
	select {
	case msg := <-mockHandler.received:
		if msg.Type != MessageTypeSctpReceive {
			t.Errorf("expected type %s, got %s", MessageTypeSctpReceive, msg.Type)
		}
		payload := msg.Payload.(ReceiveMessage)
		if payload.Stream != 1 || payload.Ppid != 1234 || string(payload.Data) != string(testData) {
			t.Errorf("unexpected payload: %+v", payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for SCTP message from server")
	}

	// 5. Test Sending from Client to Server
	t.Log("Sending message from client to server")
	clientTask.Send(runtime.Message{
		Type: MessageTypeSctpSend,
		Payload: SendMessage{
			Stream: 2,
			Ppid:   5678,
			Data:   []byte("hello sctp server"),
		},
	})

	recvBuf := make([]byte, 1024)
	serverConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	var n int
	var rinfo *sctp.SndRcvInfo
	for {
		n, rinfo, err = serverConn.SCTPRead(recvBuf)
		if err != nil {
			t.Fatalf("failed to read from client: %v", err)
		}
		if rinfo != nil {
			break
		}
		t.Logf("Skipping non-data message (notification), read %d bytes", n)
	}

	if string(recvBuf[:n]) != "hello sctp server" || rinfo.Stream != 2 || rinfo.PPID != 5678 {
		t.Errorf("unexpected data from client: %s, stream: %d, ppid: %d", string(recvBuf[:n]), rinfo.Stream, rinfo.PPID)
	}

	cancel()
	<-errCh
}
