package udp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
)

// mockTargetTaskHandler simply stores received messages for assertions.
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

func TestUdpServerTask(t *testing.T) {
	logger := logging.New("test")

	// Create a mock target task
	mockHandler := newMockTargetTaskHandler()
	targetTask := runtime.NewTask("target", logger, mockHandler, 10)

	// Listen on an ephemeral port locally
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve UDP address: %v", err)
	}

	udpHandler := NewServerTaskHandler(addr, targetTask, logger)
	udpTask := runtime.NewTask("udp_server", logger, udpHandler, 10)

	// Create a group and run it
	group := runtime.NewGroup(logger, targetTask, udpTask)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- group.Run(ctx)
	}()

	// Wait for servers to start
	time.Sleep(100 * time.Millisecond)

	// Ensure the server got an address
	if udpHandler.conn == nil {
		t.Fatal("UDP server connection is nil")
	}
	serverAddr := udpHandler.conn.LocalAddr().(*net.UDPAddr)

	// Send a test packet
	clientConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("failed to dial UDP server: %v", err)
	}
	defer clientConn.Close()

	testData := []byte("hello UERANSIM")
	if _, err := clientConn.Write(testData); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}

	// Verify the message was received by the target task
	select {
	case msg := <-mockHandler.received:
		if msg.Type != MessageTypeUdpReceive {
			t.Errorf("expected message type %q, got %q", MessageTypeUdpReceive, msg.Type)
		}
		payload, ok := msg.Payload.(ReceiveMessage)
		if !ok {
			t.Fatalf("expected payload type ReceiveMessage, got %T", msg.Payload)
		}
		if string(payload.Data) != string(testData) {
			t.Errorf("expected payload data %q, got %q", string(testData), string(payload.Data))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for UDP message")
	}

	// Test Send method
	// Send a packet back to the client
	clientLocalAddr := clientConn.LocalAddr().(*net.UDPAddr)
	replyData := []byte("reply from UERANSIM")
	if err := udpHandler.Send(clientLocalAddr, replyData); err != nil {
		t.Fatalf("failed to send reply: %v", err)
	}

	// Read reply on client side
	replyBuf := make([]byte, 1024)
	clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _, err := clientConn.ReadFromUDP(replyBuf)
	if err != nil {
		t.Fatalf("failed to read reply: %v", err)
	}
	if string(replyBuf[:n]) != string(replyData) {
		t.Errorf("expected reply %q, got %q", string(replyData), string(replyBuf[:n]))
	}

	// Shutdown
	cancel()

	// Wait for group to exit
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("group Run returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for group to shutdown")
	}
}
