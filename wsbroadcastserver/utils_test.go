// Copyright 2021-2026, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package wsbroadcastserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type eofWithDataReader struct {
	data string
}

func (r *eofWithDataReader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func TestChainedReaderEOF(t *testing.T) {
	reader := (&chainedReader{}).add(&eofWithDataReader{data: "early"}).add(strings.NewReader("frame"))
	data, err := io.ReadAll(reader)
	if err != nil || string(data) != "earlyframe" {
		t.Fatalf("ReadAll = %q, %v; want earlyframe, nil", data, err)
	}
	if n, err := reader.Read(make([]byte, 1)); n != 0 || err != io.EOF {
		t.Fatalf("exhausted reader = %d, %v; want 0, EOF", n, err)
	}
}

func TestEmptyChainedReaderEOF(t *testing.T) {
	if n, err := (&chainedReader{}).Read(make([]byte, 1)); n != 0 || err != io.EOF {
		t.Fatalf("empty reader = %d, %v; want 0, EOF", n, err)
	}
}

func TestClientMessageSizeLimit(t *testing.T) {
	for _, size := range []int{maxClientMessageSize, maxClientMessageSize + 1} {
		var wire bytes.Buffer
		if err := wsutil.WriteClientBinary(&wire, bytes.Repeat([]byte{'a'}, size)); err != nil {
			t.Fatal(err)
		}
		conn, peer := net.Pipe()
		data, _, err := ReadData(context.Background(), conn, &wire, time.Second, ws.StateServerSide, false, nil)
		_ = conn.Close()
		_ = peer.Close()
		if size > maxClientMessageSize {
			if !errors.Is(err, errClientMessageTooLarge) || data != nil {
				t.Fatalf("oversize message: data length %d, error %v", len(data), err)
			}
		} else if err != nil || len(data) != size {
			t.Fatalf("allowed message: data length %d, error %v", len(data), err)
		}
	}
}
