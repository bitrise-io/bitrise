package rde

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestBridgeConn_CopiesBothDirections(t *testing.T) {
	// clientSide <-> a is the "local" leg; b <-> serverSide the "remote" leg.
	// bridgeConn(a, b) must shuttle bytes across in both directions.
	clientSide, a := tcpConnPair(t)
	b, serverSide := tcpConnPair(t)
	defer closeConns(clientSide, a, b, serverSide)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bridgeConn(ctx, a, b)

	if _, err := clientSide.Write([]byte("ping")); err != nil {
		t.Fatalf("write client->server: %v", err)
	}
	readExpect(t, serverSide, "ping")

	if _, err := serverSide.Write([]byte("pong")); err != nil {
		t.Fatalf("write server->client: %v", err)
	}
	readExpect(t, clientSide, "pong")
}

func TestBridgeConn_DrainsResponseAfterClientHalfClose(t *testing.T) {
	// The property PR #70 fixed: a half-close in one direction must NOT tear
	// the other direction down. The client finishes its request and half-closes
	// its write; the server's reply must still reach the client before the
	// bridge returns.
	clientSide, a := tcpConnPair(t)
	b, serverSide := tcpConnPair(t)
	defer closeConns(clientSide, a, b, serverSide)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() { bridgeConn(ctx, a, b); close(done) }()

	// Client sends its request, then half-closes: the a->b copy sees EOF and
	// half-closes b, so the server observes end-of-request — but the b->a copy
	// stays open.
	if _, err := clientSide.Write([]byte("REQ")); err != nil {
		t.Fatalf("write request: %v", err)
	}
	readExpect(t, serverSide, "REQ")
	closeWrite(t, clientSide)
	// Server replies after the request side is closed; it must still arrive.
	if _, err := serverSide.Write([]byte("RESP")); err != nil {
		t.Fatalf("write response: %v", err)
	}
	readExpect(t, clientSide, "RESP")

	// Server closes; now both directions are done and the bridge returns.
	_ = serverSide.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("bridgeConn did not return after both directions closed")
	}
}

func TestBridgeConn_ReturnsOnContextCancel(t *testing.T) {
	clientSide, a := tcpConnPair(t)
	b, serverSide := tcpConnPair(t)
	defer closeConns(clientSide, a, b, serverSide)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() { bridgeConn(ctx, a, b); close(done) }()

	// No data flows and neither peer closes; only cancellation ends the bridge
	// (it must not wait for both copies in that case).
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("bridgeConn did not return after context cancel")
	}
}

func TestForwardLocal_AutoPicksLoopbackPortAndStopsOnCancel(t *testing.T) {
	// The zero-value client is safe here: we never dial a connection into the
	// listener, so forwardConn (which would use c.client) is never reached.
	// This exercises the listener lifecycle only — auto-pick, onReady, and a
	// clean (nil) return when ctx is cancelled.
	c := &sshClient{}
	ctx, cancel := context.WithCancel(context.Background())

	readyCh := make(chan string, 1)
	onReady := func(localAddr string) { readyCh <- localAddr }

	errCh := make(chan error, 1)
	go func() { errCh <- c.forwardLocal(ctx, 0, "127.0.0.1:1", onReady) }()

	var localAddr string
	select {
	case localAddr = <-readyCh:
	case <-time.After(2 * time.Second):
		t.Fatal("onReady was not called")
	}

	host, portStr, err := net.SplitHostPort(localAddr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", localAddr, err)
	}
	if host != "127.0.0.1" {
		t.Errorf("bind host = %q, want 127.0.0.1 (loopback only)", host)
	}
	if p, err := strconv.Atoi(portStr); err != nil || p == 0 {
		t.Errorf("auto-picked port = %q, want a concrete non-zero port", portStr)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("forwardLocal returned %v, want nil on cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("forwardLocal did not return after cancellation")
	}
}

// tcpConnPair returns the two ends of a real loopback TCP connection. Using
// actual sockets (rather than net.Pipe) exercises the same Read/Write/Close and
// half-close (CloseWrite) semantics bridgeConn relies on in production.
func tcpConnPair(t *testing.T) (dialed, accepted net.Conn) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	type result struct {
		conn net.Conn
		err  error
	}
	acceptCh := make(chan result, 1)
	go func() {
		conn, err := ln.Accept()
		acceptCh <- result{conn, err}
	}()

	dialed, err = net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	res := <-acceptCh
	if res.err != nil {
		t.Fatalf("accept: %v", res.err)
	}
	return dialed, res.conn
}

func closeConns(conns ...net.Conn) {
	for _, c := range conns {
		if c != nil {
			_ = c.Close()
		}
	}
}

func closeWrite(t *testing.T, c net.Conn) {
	t.Helper()
	cw, ok := c.(interface{ CloseWrite() error })
	if !ok {
		t.Fatalf("conn %T does not support CloseWrite", c)
	}
	if err := cw.CloseWrite(); err != nil {
		t.Fatalf("CloseWrite: %v", err)
	}
}

func readExpect(t *testing.T, c net.Conn, want string) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, len(want))
	if _, err := io.ReadFull(c, buf); err != nil {
		t.Fatalf("read %q: %v", want, err)
	}
	if got := string(buf); got != want {
		t.Fatalf("read = %q, want %q", got, want)
	}
}
