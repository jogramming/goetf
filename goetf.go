package goetf

type Marshaler interface {
	MarshalETF(end *Encoder) error
}

type Unmarshaler interface {
	UnmarshalETF(dec *Decoder, k string) error
}

type ListUnmarshaler interface {
	UnmarshalETFList(dec *Decoder) error
}

var _ ListUnmarshaler = (ListUnmarshaler)(nil)

type ListUnmarshalerFunc func(dec *Decoder) error

func (l ListUnmarshalerFunc) UnmarshalETFList(dec *Decoder) error {
	return l(dec)
}

// Erlang external term tags.
type TermTag byte

const (
	ETTSmallAtomUTF8 TermTag = TermTag(119)
	ETTAtomUTF8      TermTag = TermTag(118)
	ETTFun           TermTag = TermTag(117)
	ETTMap           TermTag = TermTag(116)
	ETTSmallAtom     TermTag = TermTag(115) // marked as deprecated in the spec
	ETTNewRef        TermTag = TermTag(114)
	ETTExport        TermTag = TermTag(113)
	ETTNewFun        TermTag = TermTag(112)
	ETTLargeBig      TermTag = TermTag(111)
	ETTSmallBig      TermTag = TermTag(110)
	ETTBinary        TermTag = TermTag(109)
	ETTList          TermTag = TermTag(108)
	ETTString        TermTag = TermTag(107)
	ETTNil           TermTag = TermTag(106)
	ETTLargeTuple    TermTag = TermTag(105)
	ETTSmallTuple    TermTag = TermTag(104)
	ETTPid           TermTag = TermTag(103)
	ETTPort          TermTag = TermTag(102)
	ETTRef           TermTag = TermTag(101)
	ETTAtom          TermTag = TermTag(100) // marked as deprecated in the spec
	ETTFloat         TermTag = TermTag(99)
	ETTInteger       TermTag = TermTag(98)
	ETTSmallInteger  TermTag = TermTag(97)
	ETTCacheRef      TermTag = TermTag(82)
	ETTNewCache      TermTag = TermTag(78)
	ETTBitBinary     TermTag = TermTag(77)
	ETTNewFloat      TermTag = TermTag(70)
	ETTCachedAtom    TermTag = TermTag(67)
)

func (t TermTag) String() string {
	switch t {
	case ETTAtom:
		return "ETTAtom"
	case ETTAtomUTF8:
		return "ETTAtomUTF8"
	case ETTBinary:
		return "ETTBinary"
	case ETTBitBinary:
		return "ETTBitBinary"
	case ETTCachedAtom:
		return "ETTCachedAtom"
	case ETTCacheRef:
		return "ETTCacheRef"
	case ETTExport:
		return "ETTExport"
	case ETTFloat:
		return "ETTFloat"
	case ETTFun:
		return "ETTFun"
	case ETTInteger:
		return "ETTInteger"
	case ETTLargeBig:
		return "ETTLargeBig"
	case ETTLargeTuple:
		return "ETTLargeTuple"
	case ETTList:
		return "ETTList"
	case ETTNewCache:
		return "ETTNewCache"
	case ETTNewFloat:
		return "ETTNewFloat"
	case ETTNewFun:
		return "ETTNewFun"
	case ETTNewRef:
		return "ETTNewRef"
	case ETTNil:
		return "ETTNil"
	case ETTPid:
		return "ETTPid"
	case ETTPort:
		return "ETTPort"
	case ETTRef:
		return "ETTRef"
	case ETTSmallAtom:
		return "ETTSmallAtom"
	case ETTSmallAtomUTF8:
		return "ETTSmallAtomUTF8"
	case ETTSmallBig:
		return "ETTSmallBig"
	case ETTSmallInteger:
		return "ETTSmallInteger"
	case ETTSmallTuple:
		return "ETTSmallTuple"
	case ETTString:
		return "ETTString"
	case ETTMap:
		return "ETTMap"
	}

	return "Unknown"
}
