package protocol

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/sagernet/quic-go/quicvarint"
)

func TestWriteAndReadTCPResponse(t *testing.T) {
	payload := []byte("test-payload-data")
	buffer := WriteTCPResponse(true, "OK", payload)
	defer buffer.Release()

	raw := append([]byte(nil), buffer.Bytes()...)

	// 1. Test via standard protocol reader
	r := bytes.NewReader(raw)
	ok, msg, err := ReadTCPResponse(r)
	if err != nil {
		t.Fatalf("ReadTCPResponse failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if msg != "OK" {
		t.Fatalf("expected msg=OK, got %s", msg)
	}

	remaining := make([]byte, len(payload))
	n, err := r.Read(remaining)
	if err != nil || n != len(payload) || string(remaining) != string(payload) {
		t.Fatalf("payload mismatch: %v, %s", err, string(remaining))
	}

	// 2. Strict byte-by-byte verification of full padding content
	r2 := bytes.NewReader(raw)
	var status [1]byte
	if _, err := io.ReadFull(r2, status[:]); err != nil || status[0] != 0 {
		t.Fatalf("status byte invalid: %v, status=%d", err, status[0])
	}
	bReader := quicvarint.NewReader(r2)
	msgLen, err := quicvarint.Read(bReader)
	if err != nil {
		t.Fatalf("read msgLen failed: %v", err)
	}
	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(r2, msgBytes); err != nil || string(msgBytes) != "OK" {
		t.Fatalf("read msgBytes failed: %v, msg=%s", err, string(msgBytes))
	}
	paddingLen, err := quicvarint.Read(bReader)
	if err != nil {
		t.Fatalf("read paddingLen failed: %v", err)
	}
	if paddingLen > 0 {
		padBytes := make([]byte, paddingLen)
		if _, err := io.ReadFull(r2, padBytes); err != nil {
			t.Fatalf("read full padding failed: %v", err)
		}
		// Verify every single byte in the padding is a valid padding char (not 0x00 / uninitialized heap)
		for i, c := range padBytes {
			if !strings.ContainsRune(paddingChars, rune(c)) {
				t.Fatalf("padding byte at index %d is '%c' (0x%02x), not in valid paddingChars", i, c, c)
			}
		}
	}
	gotPayload := make([]byte, len(payload))
	if _, err := io.ReadFull(r2, gotPayload); err != nil || !bytes.Equal(gotPayload, payload) {
		t.Fatalf("payload mismatch after full padding: %v", err)
	}
	if r2.Len() != 0 {
		t.Fatalf("unexpected trailing bytes: %d", r2.Len())
	}
}
