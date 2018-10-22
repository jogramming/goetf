package goetf

import (
	"bytes"
	"testing"
)

func newTestEncoder() (*Encoder, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	return enc, buf
}

func TestEncodeAtom(t *testing.T) {
	expectedOutput := []byte("\x76\x00\x0bHello world")

	enc, b := newTestEncoder()

	err := enc.WriteAtomUTF8("Hello world")
	if err != nil {
		t.Error(err)
		return
	}

	result := b.Bytes()
	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("Unexpected output, got %v, wanted: %v", result, expectedOutput)
	}
}
