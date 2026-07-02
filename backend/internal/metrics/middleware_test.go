package metrics

import (
	"bufio"
	"net"
	"net/http"
	"testing"
)

type hijackableWriter struct {
	http.ResponseWriter
	hijacked bool
}

func (w *hijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	server, client := net.Pipe()
	_ = client.Close()
	return server, bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)), nil
}

func TestStatusRecorderSupportsHijacker(t *testing.T) {
	base := &hijackableWriter{}
	rec := &statusRecorder{ResponseWriter: base, status: http.StatusOK}

	conn, _, err := rec.Hijack()
	if err != nil {
		t.Fatalf("hijack: %v", err)
	}
	_ = conn.Close()
	if !base.hijacked {
		t.Fatal("expected wrapped response writer to be hijacked")
	}
}
