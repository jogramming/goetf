package goetf

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
)

type Decoder struct {
	reader    io.Reader
	bufReader *bufio.Reader
	cursor    int

	// appends the fields when we recurse into maps
	fieldStack []string

	buf []byte
}

func NewDecoder(r io.Reader, bufSize int) *Decoder {
	dec := &Decoder{
		reader:    r,
		bufReader: bufio.NewReaderSize(r, bufSize),
	}

	return dec
}

// ReadVersion reads the version header for messages (only there is the distribution header was not)
func (dec *Decoder) ReadVersion(out *byte) error {
	b, err := dec.readNextByte()
	if err != nil {
		return err
	}

	*out = b
	return nil
}

// ReadAnyAtom reads the next atom, (one of: ATOM, ATOM_UTF8, SMALL_ATOM, SMALL_ATOM_UTF8, or STRING)
// otherwise will return an error
func (dec *Decoder) ReadAnyAtom(output *string) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	l := uint16(0)

	switch tag {
	case ETTAtom, ETTAtomUTF8, ETTString:
		l, err = dec.readUint16()
	case ETTSmallAtom, ETTSmallAtomUTF8:
		var b uint8
		b, err = dec.readNextByte()
		l = uint16(b)
	default:
		return dec.InvalidTermTag(tag, "ReadAnyAtom")
	}

	if err != nil {
		return err
	}

	err = dec.readIntoBuf(int(l))
	if err != nil {
		return err
	}

	*output = string(dec.buf)
	return nil
}

// ReadAnyBool reads the next (SMALL_)ATOM(_UTF8) and returns true if its "true", false otherwise
func (dec *Decoder) ReadAnyBool(output *bool) error {

	var str string
	err := dec.ReadAnyAtom(&str)
	if err != nil {
		return err
	}

	if str == "true" {
		*output = true
	}

	*output = false
	return nil
}

// ReadAnyString reads the next string, (one of: ATOM, ATOM_UTF8, SMALL_ATOM, SMALL_ATOM_UTF8, STRING or BINARY)
// otherwise will return an error
func (dec *Decoder) ReadAnyString(output *string) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	l := uint32(0)

	switch tag {
	case ETTAtom, ETTAtomUTF8, ETTString:
		var l16 uint16
		l16, err = dec.readUint16()
		l = uint32(l16)
	case ETTSmallAtom, ETTSmallAtomUTF8:
		var b uint8
		b, err = dec.readNextByte()
		l = uint32(b)
	case ETTBinary:
		var l32 uint32
		l32, err = dec.readUint32()
		l = l32
	default:
		return dec.InvalidTermTag(tag, "ReadString")
	}

	if err != nil {
		return err
	}

	err = dec.readIntoBuf(int(l))
	if err != nil {
		return err
	}

	*output = string(dec.buf)
	return nil
}

// ReadAnyInt32 reads the next integer, works with either INTEGER or SMALL_INTEGER, throws an error on others
func (dec *Decoder) ReadAnyInt32(dst *int32) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	switch tag {
	case ETTInteger:
		i, err := dec.readUint32()
		if err != nil {
			return err
		}
		*dst = int32(i)
	case ETTSmallInteger:
		b, err := dec.readNextByte()
		if err != nil {
			return err
		}

		*dst = int32(b)
	default:
		return dec.InvalidTermTag(tag, "ReadAnyInt32")
	}

	return nil
}

// ReadAnyInt64 reads the next integer, works with either INTEGER or SMALL_INTEGER and also with SMALL_BIG and LARGE_BIG if it can fit into an int64, throws an error on otherwise
func (dec *Decoder) ReadAnyInt64(dst *int64) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	switch tag {
	case ETTInteger:
		i, err := dec.readUint32()
		if err != nil {
			return err
		}
		*dst = int64(i)
	case ETTSmallInteger:
		b, err := dec.readNextByte()
		if err != nil {
			return err
		}

		*dst = int64(b)
	case ETTSmallBig:
		l, err := dec.readNextByte()
		if err != nil {
			return err
		}

		sign, err := dec.readNextByte()
		if err != nil {
			return err
		}

		i, err := dec.readBigIntInt64(int(l), sign)
		if err != nil {
			return err
		}
		*dst = i
	case ETTLargeBig:
		l, err := dec.readUint32()
		if err != nil {
			return err
		}

		sign, err := dec.readNextByte()
		if err != nil {
			return err
		}

		i, err := dec.readBigIntInt64(int(l), sign)
		if err != nil {
			return err
		}

		*dst = i
	default:
		return dec.InvalidTermTag(tag, "ReadAnyInt64")
	}

	return nil
}

// readBigIntInt64 reads a SMALL_BIG or LARGE_BIG into a int64, will only work on numbers that are smaller than 8 bytes
func (dec *Decoder) readBigIntInt64(l int, sign byte) (int64, error) {
	err := dec.readIntoBuf(l)
	if err != nil {
		return 0, err
	}

	if l <= 8 {
		return dec.bigIntIntoInt64(dec.buf, sign), nil
	}

	// fallback to math/big
	hsize := l >> 1
	for i := 0; i < hsize; i++ {
		dec.buf[i], dec.buf[l-i-1] = dec.buf[l-i-1], dec.buf[i]
	}

	v := new(big.Int).SetBytes(dec.buf)
	if sign != 0 {
		v = v.Neg(v)
	}

	// try int and int64
	if v.IsInt64() {
		return v.Int64(), nil
	}

	return 0, errors.New("Cannot convert big int into int64, too big")
}

// lookup table of precomputed powers of 256
// for fast SMALL_BIG and LARGE_BIG decoding into int64
var i64Powers = []int64{
	int64(math.Pow(256, 0)),
	int64(math.Pow(256, 1)),
	int64(math.Pow(256, 2)),
	int64(math.Pow(256, 3)),
	int64(math.Pow(256, 4)),
	int64(math.Pow(256, 5)),
	int64(math.Pow(256, 6)),
	int64(math.Pow(256, 7)),
}

// bigIntIntoInt64 performs a fast decode without using math/big (max 64bit integers)
func (d *Decoder) bigIntIntoInt64(b []byte, sign byte) int64 {
	if len(b) > 8 {
		return 0
	}

	var result int64
	for i := 0; i < len(b); i++ {
		result += int64(b[i]) * i64Powers[i]
	}
	if sign != 0 {
		result = -result
	}
	return result
}

// ReadAnyFloat64 reads the next NEW_FLOAT or FLOAT
func (dec *Decoder) ReadAnyFloat64(dst *float64) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	switch tag {
	case ETTFloat:
		err := dec.readIntoBuf(31)
		if err != nil {
			return err
		}

		_, err = fmt.Sscanf(string(dec.buf), "%f", &dst)
		if err != nil {
			return err
		}
	case ETTNewFloat:
		err := dec.readIntoBuf(8)
		if err != nil {
			return err
		}

		*dst = math.Float64frombits(binary.BigEndian.Uint64(dec.buf))
	default:
		err := dec.InvalidTermTag(tag, "ReadAnyFloat64")
		dec.skipAnyWithTag(tag)
		return err
	}

	return nil
}

func (dec *Decoder) ReadList(unmarshaler ListUnmarshaler) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	switch tag {
	case ETTList:
		l, err := dec.readUint32()
		if err != nil {
			return err
		}

		for i := uint32(0); i < l; i++ {
			lErr := unmarshaler.UnmarshalETFList(dec)
			if lErr != nil {
				err = lErr
			}
		}

		// skip the tail tag
		dec.skipAny()

		return err
	default:
		err := dec.InvalidTermTag(tag, "ReadList")
		dec.skipAnyWithTag(tag)
		return err
	}
}

func (dec *Decoder) ReadMapToUnmarshaler(unmarshaler Unmarshaler) error {
	tag, err := dec.readNextTag()
	if err != nil {
		return err
	}

	if tag != ETTMap {
		err := dec.InvalidTermTag(tag, "ReadMapToUnmarhsaler")
		dec.skipAnyWithTag(tag)
		return err
	}

	numPairs, err := dec.readUint32()
	if err != nil {
		return err
	}

	newStack := make([]string, len(dec.fieldStack)+1)
	copy(newStack, dec.fieldStack)
	origStack := dec.fieldStack

	for i := uint32(0); i < numPairs; i++ {
		var field string
		lErr := dec.ReadAnyAtom(&field)
		if lErr != nil {
			err = lErr    // finishing interating if there was nothing major, if there was the whole stream is messed up anyways.
			dec.skipAny() // skip value field to try to maintain decoding process
			continue
		}

		newStack[len(newStack)-1] = field
		dec.fieldStack = newStack

		lErr = unmarshaler.UnmarshalETF(dec, field)
		if lErr != nil {
			err = lErr
		}
		dec.fieldStack = origStack
	}

	return err
}

func (dec *Decoder) skipAny() {
	tag, err := dec.readNextTag()
	if err != nil {
		return
	}

	dec.skipAnyWithTag(tag)
}

func (dec *Decoder) skipAnyWithTag(tag TermTag) {
	switch tag {
	case ETTNil:
		return // contains nothing

	}
}

// readNextTag reads the next term tag
func (dec *Decoder) readNextTag() (TermTag, error) {
	b, err := dec.bufReader.ReadByte()
	dec.cursor++

	return TermTag(b), err
}

// readNextByte reads the next byte
func (dec *Decoder) readNextByte() (byte, error) {
	b, err := dec.bufReader.ReadByte()
	dec.cursor++
	return b, err
}

// readUint16 reads the next 2 bytes as a uint16
func (dec *Decoder) readUint16() (uint16, error) {
	err := dec.readIntoBuf(2)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(dec.buf), nil
}

// readUint32 reads the next 4 bytes as a uint16
func (dec *Decoder) readUint32() (uint32, error) {
	err := dec.readIntoBuf(4)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(dec.buf), nil
}

func (dec *Decoder) readIntoBuf(l int) error {
	if len(dec.buf) < l {
		if cap(dec.buf) >= l {
			dec.buf = dec.buf[:l]
		} else {
			dec.buf = make([]byte, l)
		}
	} else {
		dec.buf = dec.buf[:l]
	}

	for n := 0; n < l; {
		nr, err := dec.bufReader.Read(dec.buf)
		dec.cursor += nr
		if err != nil {
			return err
		}
		n += nr
	}

	return nil
}

func (dec *Decoder) InvalidTermTag(tag TermTag, caller string) error {
	stack := make([]string, len(dec.fieldStack))
	copy(stack, dec.fieldStack)

	return &InvalidTermTagError{
		FieldStack: stack,
		Cursor:     dec.cursor,
		TermTag:    tag,
		Caller:     caller,
	}
}

type InvalidTermTagError struct {
	FieldStack []string
	TermTag    TermTag
	Caller     string
	Cursor     int
}

func (i *InvalidTermTagError) Error() string {
	prefix := strings.Join(i.FieldStack, ": ")
	if prefix != "" {
		prefix += ": "
	}

	return fmt.Sprintf("Unexpected term tag at %d in %s: %s", i.Cursor, i.Caller, i.TermTag.String())
}
