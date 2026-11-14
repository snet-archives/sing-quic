package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteAndReadTCPResponse(t *testing.T) {
	payload := []byte("test-payload-data")
	buffer := WriteTCPResponse(true, "OK", payload)
	defer buffer.Release()

	raw := append([]byte(nil), buffer.Bytes()...)

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

	// Verify that padding region in raw is not zeroed out and consists of valid padding characters
	// Structure: [status:1][msgLen:1][msg:2][padLen:1-2][padding...][payload]
	padIndex := 1 + 1 + 2 + 2 // approx start of padding
	if len(raw) > padIndex+50 {
		padSample := string(raw[padIndex : padIndex+50])
		for _, c := range padSample {
			if !strings.ContainsRune(paddingChars, c) {
				t.Fatalf("padding byte %c is not in valid paddingChars", c)
			}
		}
	}
}

