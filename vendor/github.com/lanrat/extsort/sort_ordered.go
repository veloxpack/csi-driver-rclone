package extsort

import (
	"cmp"
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"unsafe"
)

// OrderedSorter provides external sorting for types that implement cmp.Ordered.
// It embeds GenericSorter and serializes records with a compact binary codec
// chosen once for the underlying kind of T.
type OrderedSorter[T cmp.Ordered] struct {
	GenericSorter[T]
}

// errOrderedDecode reports a record that is not a valid encoding of the sorted type.
var errOrderedDecode = errors.New("extsort: invalid encoding of an ordered value")

// orderedCodec returns serialization functions for T, chosen by its underlying kind, so
// named types such as `type ID int64` use the codec of their underlying type. Integers are
// stored as varints, floats as their exact IEEE 754 bits (so NaN and -0 survive), and
// strings as their bytes.
func orderedCodec[T cmp.Ordered]() (FromBytesGeneric[T], ToBytesGeneric[T]) {
	fromBytes, appendBytes := orderedAppendCodec[T]()
	return fromBytes, toBytesFunc(appendBytes)
}

// orderedAppendCodec returns the codec of orderedCodec, with the encoder in append form.
//
// The codecs read and write T through a pointer to its underlying type. The kind check
// guarantees the two types share a memory layout, which makes the conversion valid.
func orderedAppendCodec[T cmp.Ordered]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	switch kind := reflect.TypeFor[T]().Kind(); kind {
	case reflect.Int:
		return signedCodec[T, int]()
	case reflect.Int8:
		return signedCodec[T, int8]()
	case reflect.Int16:
		return signedCodec[T, int16]()
	case reflect.Int32:
		return signedCodec[T, int32]()
	case reflect.Int64:
		return signedCodec[T, int64]()
	case reflect.Uint:
		return unsignedCodec[T, uint]()
	case reflect.Uint8:
		return unsignedCodec[T, uint8]()
	case reflect.Uint16:
		return unsignedCodec[T, uint16]()
	case reflect.Uint32:
		return unsignedCodec[T, uint32]()
	case reflect.Uint64:
		return unsignedCodec[T, uint64]()
	case reflect.Uintptr:
		return unsignedCodec[T, uintptr]()
	case reflect.Float32:
		return float32Codec[T]()
	case reflect.Float64:
		return float64Codec[T]()
	case reflect.String:
		return stringCodec[T]()
	default:
		panic("extsort: unsupported cmp.Ordered kind " + kind.String())
	}
}

// signedCodec stores a T whose underlying type is I as a zig-zag varint.
func signedCodec[T cmp.Ordered, I int | int8 | int16 | int32 | int64]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	fromBytes := func(d []byte) (T, error) {
		var v T
		x, n := binary.Varint(d)
		if n <= 0 || n != len(d) || int64(I(x)) != x {
			return v, errOrderedDecode
		}
		*(*I)(unsafe.Pointer(&v)) = I(x)
		return v, nil
	}
	appendBytes := func(dst []byte, v T) []byte {
		return binary.AppendVarint(dst, int64(*(*I)(unsafe.Pointer(&v))))
	}
	return fromBytes, appendBytes
}

// unsignedCodec stores a T whose underlying type is U as a varint.
func unsignedCodec[T cmp.Ordered, U uint | uint8 | uint16 | uint32 | uint64 | uintptr]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	fromBytes := func(d []byte) (T, error) {
		var v T
		x, n := binary.Uvarint(d)
		if n <= 0 || n != len(d) || uint64(U(x)) != x {
			return v, errOrderedDecode
		}
		*(*U)(unsafe.Pointer(&v)) = U(x)
		return v, nil
	}
	appendBytes := func(dst []byte, v T) []byte {
		return binary.AppendUvarint(dst, uint64(*(*U)(unsafe.Pointer(&v))))
	}
	return fromBytes, appendBytes
}

// float32Codec stores a T whose underlying type is float32 as its 4 IEEE 754 bytes.
func float32Codec[T cmp.Ordered]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	fromBytes := func(d []byte) (T, error) {
		var v T
		if len(d) != 4 {
			return v, errOrderedDecode
		}
		*(*float32)(unsafe.Pointer(&v)) = math.Float32frombits(binary.BigEndian.Uint32(d))
		return v, nil
	}
	appendBytes := func(dst []byte, v T) []byte {
		return binary.BigEndian.AppendUint32(dst, math.Float32bits(*(*float32)(unsafe.Pointer(&v))))
	}
	return fromBytes, appendBytes
}

// float64Codec stores a T whose underlying type is float64 as its 8 IEEE 754 bytes.
func float64Codec[T cmp.Ordered]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	fromBytes := func(d []byte) (T, error) {
		var v T
		if len(d) != 8 {
			return v, errOrderedDecode
		}
		*(*float64)(unsafe.Pointer(&v)) = math.Float64frombits(binary.BigEndian.Uint64(d))
		return v, nil
	}
	appendBytes := func(dst []byte, v T) []byte {
		return binary.BigEndian.AppendUint64(dst, math.Float64bits(*(*float64)(unsafe.Pointer(&v))))
	}
	return fromBytes, appendBytes
}

// stringCodec stores a T whose underlying type is string as its bytes.
func stringCodec[T cmp.Ordered]() (FromBytesGeneric[T], appendBytesFunc[T]) {
	fromBytes := func(d []byte) (T, error) {
		var v T
		*(*string)(unsafe.Pointer(&v)) = string(d)
		return v, nil
	}
	appendBytes := func(dst []byte, v T) []byte {
		return append(dst, *(*string)(unsafe.Pointer(&v))...)
	}
	return fromBytes, appendBytes
}

// Ordered performs external sorting on a channel of cmp.Ordered types.
// It returns the sorter instance, output channel with sorted results, and error channel.
// Records are serialized with a compact binary codec for T's underlying kind and
// compared with cmp.Compare, which orders NaN before all other floats.
//
// IMPORTANT: The input channel MUST be closed to signal the end of data.
// Sort() will continue reading from the input channel until it is closed.
func Ordered[T cmp.Ordered](input <-chan T, config *Config) (*OrderedSorter[T], <-chan T, <-chan error) {
	fromBytes, appendBytes := orderedAppendCodec[T]()
	s, output, errChan := Generic(input, fromBytes, toBytesFunc(appendBytes), cmp.Compare, config)
	s.useBuiltinCodec(appendBytes)
	return &OrderedSorter[T]{GenericSorter: *s}, output, errChan
}

// OrderedMock performs external sorting with a mock implementation that limits
// the number of items to sort (useful for testing). Takes the same parameters as
// Ordered plus n which limits the number of items processed.
func OrderedMock[T cmp.Ordered](input <-chan T, config *Config, n int) (*OrderedSorter[T], <-chan T, <-chan error) {
	fromBytes, appendBytes := orderedAppendCodec[T]()
	s, output, errChan := MockGeneric(input, fromBytes, toBytesFunc(appendBytes), cmp.Compare, config, n)
	s.useBuiltinCodec(appendBytes)
	return &OrderedSorter[T]{GenericSorter: *s}, output, errChan
}
