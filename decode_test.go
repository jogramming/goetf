package goetf

import (
	"bytes"
	"testing"
)

func newTestDecoder(source []byte) *Decoder {
	b := bytes.NewBuffer(source)
	return NewDecoder(b, 0xff)
}

func TestDecodeAtom(t *testing.T) {
	input := []byte("\x73\x0bHello world")
	expectedOutput := "Hello world"

	dec := newTestDecoder(input)
	var s string
	err := dec.ReadAtom(&s)
	if err != nil {
		t.Error(err)
	}

	if s != expectedOutput {
		t.Error("Unexpected output: ", s, expectedOutput)
	}
}
