package goetf

import (
	"encoding/binary"
	"io"
)

type Encoder struct {
	writer io.Writer

	// appends the fields when we recurse into maps
	fieldStack []string

	buf []byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

func (enc *Encoder) WriteVersion() error {
	_, err := enc.writer.Write([]byte{131})
	return err
}

func (enc *Encoder) WriteAtomUTF8(atom string) error {
	b := make([]byte, 3)
	b[0] = byte(ETTAtomUTF8)
	binary.BigEndian.PutUint16(b[1:], uint16(len(atom)))

	_, err := enc.writer.Write(b)
	if err != nil {
		return err
	}

	_, err = enc.writer.Write([]byte(atom))
	return err
}

func (enc *Encoder) WriteBinaryString(str string) error {
	b := make([]byte, 5)
	b[0] = byte(ETTBinary)
	binary.BigEndian.PutUint32(b[1:], uint32(len(str)))

	_, err := enc.writer.Write(b)
	if err != nil {
		return err
	}

	_, err = enc.writer.Write([]byte(str))
	return err
}
