// Package srvsvc implements the parts of the Server Service Remote Protocol
// that enumerate the shares on a server.
//
// The protocol is documented by Microsoft as [MS-SRVS], at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/4b07ff60-dada-4580-a6d8-4668bffbde6c
//
// The specification is named MS-SRVS; the RPC interface and the named pipe
// used to call it are both named srvsvc.
//
// The msrpc package encodes the PDU that contains each message. This package
// defines the interface identity (UUID, version and opnums) and the stubs,
// which are the call arguments marshalled in network data representation, or
// NDR.
package srvsvc

import (
	"encoding/binary"

	"github.com/cloudsoda/go-smb2/internal/utf16le"
)

var le = binary.LittleEndian

func roundup(x, align int) int {
	return (x + (align - 1)) &^ (align - 1)
}

const (
	// VERSION and VERSION_MINOR are the version of the srvsvc interface, 3.0.
	VERSION       = 3
	VERSION_MINOR = 0

	// OP_NET_SHARE_ENUM is the opnum of NetrShareEnum, defined in [MS-SRVS]
	// section 3.1.4.8, at
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/c4a98e7b-d416-439c-97bd-4d9f52f8ba52
	OP_NET_SHARE_ENUM = 15
)

// UUID is the srvsvc interface identifier 4b324fc8-1670-01d3-1278-5a47bf6ee188
// in wire order: the first three fields are byte swapped.
var UUID = [16]byte{
	0xc8, 0x4f, 0x32, 0x4b, 0x70, 0x16, 0xd3, 0x01,
	0x12, 0x78, 0x5a, 0x47, 0xbf, 0x6e, 0xe1, 0x88,
}

// ShareInfoLevel is the Level member of a SHARE_ENUM_STRUCT, which selects the
// SHARE_INFO_* structure that NetrShareEnum returns. The valid levels are
// listed with NetrShareEnum in [MS-SRVS] section 3.1.4.8, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/c4a98e7b-d416-439c-97bd-4d9f52f8ba52
//
// The valid levels are 0, 1, 2, 501, 502 and 503; a server fails the call with
// ERROR_INVALID_LEVEL for any other. This package supports only ShareInfoLevel0
// and ShareInfoLevel1.
type ShareInfoLevel uint32

const (
	// ShareInfoLevel0 selects SHARE_INFO_0, which contains the share name.
	ShareInfoLevel0 ShareInfoLevel = 0
	// ShareInfoLevel1 selects SHARE_INFO_1, which contains the share name,
	// type and remark.
	ShareInfoLevel1 ShareInfoLevel = 1
)

// The STYPE_* constants are the share types defined in [MS-SRVS] section
// 2.2.2.4, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/6069f8c0-c93f-43a0-a5b4-7ed447eb4b84
//
// The low bits hold the base type and the high bits hold independent flags.
const (
	STYPE_DISKTREE = 0x00000000
	STYPE_PRINTQ   = 0x00000001
	STYPE_DEVICE   = 0x00000002
	STYPE_IPC      = 0x00000003

	STYPE_TEMPORARY = 0x40000000
	STYPE_SPECIAL   = 0x80000000

	// STYPE_MASK masks a share type to its base type. [MS-SRVS] does not
	// define it; the value is the one lmshare.h in the Windows SDK defines.
	STYPE_MASK = 0x000000FF
)

// entrySize returns the size of the fixed part of one array element at level l,
// and false for a level other than ShareInfoLevel0 or ShareInfoLevel1. The
// fixed part is all that is stored in the array; the strings it points to
// follow the array.
func (l ShareInfoLevel) entrySize() (int, bool) {
	switch l {
	case ShareInfoLevel0:
		return 4, true // netname pointer
	case ShareInfoLevel1:
		return 12, true // netname pointer, type, remark pointer
	default:
		return 0, false
	}
}

// ShareInfo1 is a decoded SHARE_INFO_1, defined in [MS-SRVS] section 2.2.4.23,
// at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/fc69f110-998d-4c16-9667-514e22fdd80b
//
// Type and Comment are zero in a response at ShareInfoLevel0, which contains
// only the name.
type ShareInfo1 struct {
	Name    string
	Type    uint32
	Comment string
}

// NetShareEnumAllRequest is the NDR stub of a NetrShareEnum call, defined in
// [MS-SRVS] section 3.1.4.8, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/c4a98e7b-d416-439c-97bd-4d9f52f8ba52
type NetShareEnumAllRequest struct {
	ServerName string
	Level      ShareInfoLevel
}

// Size returns the length of the encoded stub.
func (r *NetShareEnumAllRequest) Size() int {
	off := 16 + utf16le.EncodedStringLen(r.ServerName) + 2
	off = roundup(off, 4)
	off += 24
	off += 4
	return off
}

// Encode writes the stub to b, which must be at least Size bytes long. It does
// not write a PDU header.
func (r *NetShareEnumAllRequest) Encode(b []byte) {
	// following parts will change if we use NDR64 instead of NDR

	// ServerName

	le.PutUint32(b[0:4], 0x20000) // referent ID

	count := utf16le.EncodedStringLen(r.ServerName)/2 + 1

	le.PutUint32(b[4:8], uint32(count))   // max count
	le.PutUint32(b[8:12], 0)              // offset
	le.PutUint32(b[12:16], uint32(count)) // actual count

	n := utf16le.EncodeSlice(b[16:], r.ServerName, utf16le.MapCharsNone)

	off := roundup(16+count*2, 4)

	// EncodeSlice writes neither the NUL that actual_count includes nor the
	// padding that follows it, and Encode may not assume b is zeroed.
	clear(b[16+n : off])

	// InfoStruct

	le.PutUint32(b[off:off+4], uint32(r.Level))   // Level
	le.PutUint32(b[off+4:off+8], uint32(r.Level)) // ShareInfo union switch
	le.PutUint32(b[off+8:off+12], 0x20004)        // container referent ID
	le.PutUint32(b[off+12:off+16], 0)             // container EntriesRead
	le.PutUint32(b[off+16:off+20], 0)             // container Buffer, a null pointer

	// PreferedMaximumLength

	le.PutUint32(b[off+20:off+24], 0xffffffff) // MAX_PREFERRED_LENGTH

	off += 24

	// ResumeHandle

	le.PutUint32(b[off:off+4], 0) // null pointer
}

// NetShareEnumAllResponseDecoder decodes the stub of a NetrShareEnum response:
// a SHARE_ENUM_STRUCT, defined in [MS-SRVS] section 2.2.4.38, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/79ee052e-e16b-4ec5-b4b7-e99777c26eca
//
// followed by TotalEntries, ResumeHandle and the return value. The stub of a
// response sent in several fragments is the concatenation of the stubs of all
// its fragments.
type NetShareEnumAllResponseDecoder []byte

// Level returns the level of the response, and false if c is too short to
// contain it.
func (c NetShareEnumAllResponseDecoder) Level() (ShareInfoLevel, bool) {
	if len(c) < 4 {
		return 0, false
	}
	return ShareInfoLevel(le.Uint32(c[0:4])), true
}

// Decode returns the shares in the response and the WERROR that NetrShareEnum
// returned, which is 0 if the call succeeded. ok is false, and shares is nil,
// if c is truncated, its union switch does not match its level, its entry
// count is larger than c can hold, or its level is not ShareInfoLevel0 or
// ShareInfoLevel1.
func (c NetShareEnumAllResponseDecoder) Decode() (shares []ShareInfo1, returnValue uint32, ok bool) {
	r := stubReader{b: c, ok: true}

	level := ShareInfoLevel(r.readUint32())
	unionSwitch := r.readUint32()
	if unionSwitch != uint32(level) {
		return nil, 0, false
	}
	entrySize, ok := level.entrySize()
	if !ok || !r.ok {
		return nil, 0, false
	}

	shares = []ShareInfo1{}

	containerPtr := r.readUint32()
	if containerPtr != 0 {
		count := r.readUint32()
		arrayPtr := r.readUint32()
		if arrayPtr != 0 {
			maxCount := r.readUint32()

			// Check the entry count against the stub length before allocating
			// for it.
			if !r.ok || maxCount != count || int(count) > (len(c)-r.off)/entrySize {
				return nil, 0, false
			}

			fixed := r.off
			r.off += int(count) * entrySize
			shares = make([]ShareInfo1, count)

			// The strings follow the array of fixed parts, in element order:
			// name[0], remark[0], name[1], remark[1], ...
			for i := range shares {
				entry := c[fixed+i*entrySize:]

				namePtr := le.Uint32(entry[0:4])
				var remarkPtr uint32
				if entrySize == 12 {
					shares[i].Type = le.Uint32(entry[4:8])
					remarkPtr = le.Uint32(entry[8:12])
				}

				// A string with a NULL referent ID is not present in the stub.
				if namePtr != 0 {
					shares[i].Name = r.readConformantVaryingString()
				}
				if remarkPtr != 0 {
					shares[i].Comment = r.readConformantVaryingString()
				}
			}
		} else if count != 0 {
			return nil, 0, false
		}
	}

	r.readUint32() // TotalEntries
	resumeHandlePtr := r.readUint32()
	if resumeHandlePtr != 0 {
		r.readUint32() // ResumeHandle
	}
	returnValue = r.readUint32()

	if !r.ok {
		return nil, 0, false
	}
	return shares, returnValue, true
}

// A stubReader reads NDR values from b in order, starting at off. Once a read
// extends past the end of b, ok is false and every later read returns a zero
// value.
type stubReader struct {
	b   []byte
	off int
	ok  bool
}

func (r *stubReader) readUint32() uint32 {
	if !r.ok || len(r.b)-r.off < 4 {
		r.ok = false
		return 0
	}
	v := le.Uint32(r.b[r.off:])
	r.off += 4
	return v
}

// readConformantVaryingString reads an NDR conformant and varying string and
// advances to the next 4-byte boundary. The representation is defined in C706
// section 14.3.4.2, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm#tagcjh_19_03_04_02
func (r *stubReader) readConformantVaryingString() string {
	r.readUint32() // max_count

	// The offset field is an index into the conceptual array, not a byte
	// offset, so it is not used: the actual_count elements always follow the
	// three count fields.
	r.readUint32()

	n := int(r.readUint32()) * 2 // actual_count UTF-16 code units, 2 bytes each
	if !r.ok || len(r.b)-r.off < n {
		r.ok = false
		return ""
	}

	// utf16le.Decode drops the terminating NUL that actual_count includes.
	s := utf16le.Decode(r.b[r.off:r.off+n], utf16le.MapCharsNone)
	r.off = roundup(r.off+n, 4)
	return s
}
