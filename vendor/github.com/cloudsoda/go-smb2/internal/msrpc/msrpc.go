// Package msrpc implements the client side of the connection-oriented RPC
// protocol used to call remote procedures over an SMB named pipe.
//
// A Windows server exposes services such as srvsvc and lsarpc as RPC
// interfaces, each called through the matching named pipe on the IPC$ share.
// Calling one takes three steps: negotiating which interface subsequent calls
// address (Bind, BindAckDecoder), sending a call's marshalled arguments and
// operation number in a request (Request), and decoding the reply
// (ResponseDecoder). This package encodes and decodes those PDUs. The caller
// writes requests to the pipe, and ReadResponse assembles a response from the
// bytes that a caller-supplied function reads from the pipe.
//
// The first step is presentation context negotiation. A presentation context
// pairs an abstract syntax, which is the interface being called, identified by
// its UUID and version, with a transfer syntax, which is the encoding of its
// arguments, and assigns the pair a small integer id. The client proposes one
// or more contexts in a bind, and the server accepts or rejects each one. Every
// later request contains only the context id, so the interface and the encoding
// are not repeated per call. Bind proposes exactly one context, with id 0,
// pairing an interface with NDR.
//
// Each message exchanged on the pipe is a protocol data unit, or PDU. Every PDU
// begins with a fixed header that includes the message type; in a request or a
// response, the call's arguments follow that header. The arguments are
// marshalled using network data representation, or NDR, and the marshalled
// bytes are known as the stub.
//
// The PDU layouts are defined by The Open Group in C706, DCE 1.1: Remote
// Procedure Call, chapter 12, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//
// [MS-RPCE], the Remote Procedure Call Protocol Extensions, does not restate
// them: it references C706 normatively and documents only Microsoft's
// additions, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/290c38b1-92fe-4229-91e6-4fc376610c15
//
// This package implements only the PDU framing and is independent of any
// interface. The UUID, opnums and NDR stubs of an interface belong in a package
// for that interface, and a stub is sent by setting it as Request.Stub.
package msrpc

import (
	"encoding/binary"
	"fmt"
)

var le = binary.LittleEndian

func roundup(x, align int) int {
	return (x + (align - 1)) &^ (align - 1)
}

const (
	// RPC_VERSION and RPC_VERSION_MINOR are the version of the
	// connection-oriented protocol, 5.0. C706 permits a minor version of 0 or 1, and
	// describes minor version 1 as supporting security credentials larger
	// than 1400 bytes. This package sends no credentials.
	RPC_VERSION       = 5
	RPC_VERSION_MINOR = 0

	// RPC_TYPE_REQUEST, RPC_TYPE_RESPONSE, RPC_TYPE_FAULT, RPC_TYPE_BIND and
	// RPC_TYPE_BIND_ACK are PTYPE values from the RPC Protocol Data Units table
	// in C706 chapter 12. The numbering is shared with the connectionless
	// protocol: request, response and fault are common to both, 1 and 4
	// through 10 are connectionless only, and the connection-oriented types
	// begin at 11.
	RPC_TYPE_REQUEST  = 0
	RPC_TYPE_RESPONSE = 2
	RPC_TYPE_FAULT    = 3
	RPC_TYPE_BIND     = 11
	RPC_TYPE_BIND_ACK = 12

	// RPC_PACKET_FLAG_FIRST and RPC_PACKET_FLAG_LAST are bits in the pfc_flags
	// header field, declared in C706 section 12.6.3.1. A PDU that is not
	// fragmented sets both.
	RPC_PACKET_FLAG_FIRST = 0x01
	RPC_PACKET_FLAG_LAST  = 0x02

	// NDR_VERSION is the version of the transfer syntax identified by
	// NDR_UUID. A bind writes it after NDR_UUID in a p_syntax_id_t.
	NDR_VERSION = 2

	// CommonHeaderSize is the size of the header shared by every PDU type.
	CommonHeaderSize = 16

	// StubOffset is the size of a request or response PDU header, and so the
	// offset at which the NDR stub begins.
	StubOffset = 24

	// MaxXmitFrag is the value Bind writes to both max_xmit_frag and
	// max_recv_frag. The bind_ack contains the server's own limits, which this
	// package does not read.
	MaxXmitFrag = 4280
)

// RPC_RESULT_ACCEPTANCE, RPC_RESULT_USER_REJECTION and
// RPC_RESULT_PROVIDER_REJECTION are the values of p_cont_def_result_t, the
// result a bind_ack contains for each presentation context proposed in a bind.
// They are declared in C706 section 12.6.3.1, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_03_01
const (
	RPC_RESULT_ACCEPTANCE         = 0
	RPC_RESULT_USER_REJECTION     = 1
	RPC_RESULT_PROVIDER_REJECTION = 2
)

// NDR_UUID is the transfer syntax identifier of NDR, in wire order. Its version
// is NDR_VERSION. The encoding is defined in C706 chapter 14, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
var NDR_UUID = [16]byte{
	0x04, 0x5d, 0x88, 0x8a, 0xeb, 0x1c, 0xc9, 0x11,
	0x9f, 0xe8, 0x08, 0x00, 0x2b, 0x10, 0x48, 0x60,
}

// An Encoder encodes the NDR stub of an RPC request. Size returns the number of
// bytes Encode writes, and Encode requires b to be at least that long.
type Encoder interface {
	Size() int
	Encode(b []byte)
}

// A HeaderDecoder decodes the fields shared by every PDU type, as declared in
// C706 section 12.6.3.1, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_03_01
type HeaderDecoder []byte

// IsInvalid reports whether c is too short to contain a PDU header or has a
// version other than RPC_VERSION.RPC_VERSION_MINOR. It does not check the
// packet type.
func (c HeaderDecoder) IsInvalid() bool {
	if len(c) < CommonHeaderSize {
		return true
	}
	if c.Version() != RPC_VERSION {
		return true
	}
	if c.VersionMinor() != RPC_VERSION_MINOR {
		return true
	}
	return false
}

// Version returns the rpc_vers field.
func (c HeaderDecoder) Version() uint8 {
	return c[0]
}

// VersionMinor returns the rpc_vers_minor field.
func (c HeaderDecoder) VersionMinor() uint8 {
	return c[1]
}

// PacketType returns the PTYPE field.
func (c HeaderDecoder) PacketType() uint8 {
	return c[2]
}

// PacketFlags returns the pfc_flags field.
func (c HeaderDecoder) PacketFlags() uint8 {
	return c[3]
}

// DataRepresentation returns the packed_drep field.
func (c HeaderDecoder) DataRepresentation() []byte {
	return c[4:8]
}

// FragLength returns the frag_length field, the length of the PDU in bytes.
func (c HeaderDecoder) FragLength() uint16 {
	return le.Uint16(c[8:10])
}

// AuthLength returns the auth_length field.
func (c HeaderDecoder) AuthLength() uint16 {
	return le.Uint16(c[10:12])
}

// CallId returns the call_id field.
func (c HeaderDecoder) CallId() uint32 {
	return le.Uint32(c[12:16])
}

// A Bind is a bind PDU that proposes one presentation context, with id 0: the
// interface identified by InterfaceUUID and its version as the abstract syntax,
// and NDR as the only transfer syntax. The bind PDU layout is defined in C706
// section 12.6.4.3, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_04_03
type Bind struct {
	CallId uint32
	// InterfaceUUID is in wire order: the first three fields of the UUID are
	// byte swapped.
	InterfaceUUID [16]byte
	// InterfaceVersion and InterfaceVersionMinor together make up the u_int32
	// if_version of a p_syntax_id_t: C706 places the major version in its 16
	// least significant bits and the minor version in the 16 most
	// significant.
	InterfaceVersion      uint16
	InterfaceVersionMinor uint16
}

// Size returns the length of the encoded bind PDU.
func (r *Bind) Size() int {
	return 72
}

// Encode writes the bind PDU to b, which must be at least Size bytes long.
func (r *Bind) Encode(b []byte) {
	b[0] = RPC_VERSION
	b[1] = RPC_VERSION_MINOR
	b[2] = RPC_TYPE_BIND
	b[3] = RPC_PACKET_FLAG_FIRST | RPC_PACKET_FLAG_LAST

	// order = Little-Endian, float = IEEE, char = ASCII
	b[4] = 0x10
	b[5] = 0
	b[6] = 0
	b[7] = 0

	le.PutUint16(b[8:10], 72)           // frag length
	le.PutUint16(b[10:12], 0)           // auth length
	le.PutUint32(b[12:16], r.CallId)    // call id
	le.PutUint16(b[16:18], MaxXmitFrag) // max xmit frag
	le.PutUint16(b[18:20], MaxXmitFrag) // max recv frag
	le.PutUint32(b[20:24], 0)           // assoc group
	le.PutUint32(b[24:28], 1)           // num ctx items
	le.PutUint16(b[28:30], 0)           // ctx item[1] .context id
	le.PutUint16(b[30:32], 1)           // ctx item[1] .num trans items

	copy(b[32:48], r.InterfaceUUID[:])
	le.PutUint16(b[48:50], r.InterfaceVersion)
	le.PutUint16(b[50:52], r.InterfaceVersionMinor)

	copy(b[52:68], NDR_UUID[:])
	le.PutUint32(b[68:72], NDR_VERSION)
}

// A BindAckDecoder decodes a bind_ack PDU, whose layout is defined in C706
// section 12.6.4.4, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_04_04
type BindAckDecoder struct {
	HeaderDecoder
}

// NewBindAckDecoder returns a BindAckDecoder that decodes b.
func NewBindAckDecoder(b []byte) BindAckDecoder {
	return BindAckDecoder{HeaderDecoder(b)}
}

// bindAckSecAddrOffset is the offset of the sec_addr field in a bind_ack PDU,
// after max_xmit_frag, max_recv_frag and assoc_group_id.
const bindAckSecAddrOffset = 24

// IsInvalid reports whether c is too short to contain the bind_ack fields that
// precede sec_addr, has an unsupported version, or is not a bind_ack.
func (c BindAckDecoder) IsInvalid() bool {
	if c.HeaderDecoder.IsInvalid() {
		return true
	}
	if len(c.HeaderDecoder) < bindAckSecAddrOffset {
		return true
	}
	if c.PacketType() != RPC_TYPE_BIND_ACK {
		return true
	}
	return false
}

// MaxXmitFrag returns the max_xmit_frag field.
func (c BindAckDecoder) MaxXmitFrag() uint16 {
	return le.Uint16(c.HeaderDecoder[16:18])
}

// MaxRecvFrag returns the max_recv_frag field.
func (c BindAckDecoder) MaxRecvFrag() uint16 {
	return le.Uint16(c.HeaderDecoder[18:20])
}

// AssocGroupId returns the assoc_group_id field.
func (c BindAckDecoder) AssocGroupId() uint32 {
	return le.Uint32(c.HeaderDecoder[20:24])
}

// resultListOffset returns the offset of p_result_list, or -1 if c is too short
// to contain it. The list follows the variable-length sec_addr field, a 2-byte
// length and that many bytes, and is aligned to a 4-byte boundary.
func (c BindAckDecoder) resultListOffset() int {
	start := bindAckSecAddrOffset + 2
	if len(c.HeaderDecoder) < start {
		return -1
	}
	off := roundup(start+int(le.Uint16(c.HeaderDecoder[bindAckSecAddrOffset:start])), 4)
	if len(c.HeaderDecoder) < off+4 {
		return -1
	}
	return off
}

// IsAccepted reports whether p_result_list contains a result of
// RPC_RESULT_ACCEPTANCE. A server that rejects a presentation context still
// returns a bind_ack rather than a bind_nak, so a valid bind_ack alone does not
// mean the context was accepted.
func (c BindAckDecoder) IsAccepted() bool {
	off := c.resultListOffset()
	if off < 0 {
		return false
	}

	for i := range int(c.HeaderDecoder[off]) { // n_results
		roff := off + 4 + i*24 // result_t is 24 bytes
		if len(c.HeaderDecoder) < roff+2 {
			return false
		}
		if le.Uint16(c.HeaderDecoder[roff:roff+2]) == RPC_RESULT_ACCEPTANCE {
			return true
		}
	}

	return false
}

// A Request is an unfragmented request PDU whose body is Stub. The request PDU
// layout is defined in C706 section 12.6.4.9, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_04_09
type Request struct {
	CallId uint32
	// ContextId is the id of a presentation context accepted by an earlier
	// bind. Bind proposes context id 0.
	ContextId uint16
	Opnum     uint16
	Stub      Encoder
}

// Size returns the length of the encoded request PDU.
func (r *Request) Size() int {
	return StubOffset + r.Stub.Size()
}

// Encode writes the request PDU to b, which must be at least Size bytes long.
func (r *Request) Encode(b []byte) {
	b[0] = RPC_VERSION
	b[1] = RPC_VERSION_MINOR
	b[2] = RPC_TYPE_REQUEST
	b[3] = RPC_PACKET_FLAG_FIRST | RPC_PACKET_FLAG_LAST

	// order = Little-Endian, float = IEEE, char = ASCII
	b[4] = 0x10
	b[5] = 0
	b[6] = 0
	b[7] = 0

	le.PutUint16(b[10:12], 0)           // auth length
	le.PutUint32(b[12:16], r.CallId)    // call id
	le.PutUint16(b[20:22], r.ContextId) // context id
	le.PutUint16(b[22:24], r.Opnum)     // opnum

	stubLen := r.Stub.Size()
	r.Stub.Encode(b[StubOffset : StubOffset+stubLen])

	le.PutUint16(b[8:10], uint16(StubOffset+stubLen)) // frag length
	le.PutUint32(b[16:20], uint32(stubLen))           // alloc hint
}

// A ResponseDecoder decodes a response PDU, whose layout is defined in C706
// section 12.6.4.10, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_04_10
type ResponseDecoder struct {
	HeaderDecoder
}

// NewResponseDecoder returns a ResponseDecoder that decodes b.
func NewResponseDecoder(b []byte) ResponseDecoder {
	return ResponseDecoder{HeaderDecoder(b)}
}

// IsInvalid reports whether c is too short to contain a response header or its
// frag_length, has an unsupported version, is not a response, or has
// authentication fields that do not fit in frag_length.
func (c ResponseDecoder) IsInvalid() bool {
	if c.HeaderDecoder.IsInvalid() {
		return true
	}
	if c.PacketType() != RPC_TYPE_RESPONSE {
		return true
	}
	return c.stubEnd() < 0
}

// secTrailerSize is the size of the fields between the authentication padding
// and auth_value: auth_type, auth_level, auth_pad_length, auth_reserved and
// auth_context_id.
const secTrailerSize = 8

// stubEnd returns the offset at which the stub ends, or -1 if c is shorter than
// StubOffset or than its frag_length, or if the authentication fields do not
// fit in frag_length. When auth_length is non-zero, frag_length includes an
// authentication verifier after the stub: auth_pad_length bytes of padding,
// the sec trailer and auth_length bytes of auth_value. The verifier is defined
// in C706 section 13.2.6.1, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap13.htm#tagcjh_18_02_06_01
func (c ResponseDecoder) stubEnd() int {
	if len(c.HeaderDecoder) < StubOffset {
		return -1
	}

	end := int(c.FragLength())
	if end < StubOffset || end > len(c.HeaderDecoder) {
		return -1
	}

	if authLen := int(c.AuthLength()); authLen > 0 {
		end -= authLen + secTrailerSize
		if end < StubOffset {
			return -1
		}
		end -= int(c.HeaderDecoder[end+2]) // auth_pad_length
		if end < StubOffset {
			return -1
		}
	}

	return end
}

// AllocHint returns the alloc_hint field.
func (c ResponseDecoder) AllocHint() uint32 {
	return le.Uint32(c.HeaderDecoder[16:20])
}

// ContextId returns the p_cont_id field.
func (c ResponseDecoder) ContextId() uint16 {
	return le.Uint16(c.HeaderDecoder[20:22])
}

// CancelCount returns the cancel_count field.
func (c ResponseDecoder) CancelCount() uint8 {
	return c.HeaderDecoder[22]
}

// Stub returns the stub of the response, excluding any bytes past frag_length
// and any authentication verifier, or nil if IsInvalid reports true.
func (c ResponseDecoder) Stub() []byte {
	if c.IsInvalid() {
		return nil
	}
	return c.HeaderDecoder[StubOffset:c.stubEnd()]
}

// faultStatusEnd is the offset of the end of the status field in a fault PDU.
const faultStatusEnd = 28

// A FaultDecoder decodes a fault PDU, whose layout is defined in C706 section
// 12.6.4.7, at
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm#tagcjh_17_06_04_07
type FaultDecoder struct {
	HeaderDecoder
}

// NewFaultDecoder returns a FaultDecoder that decodes b.
func NewFaultDecoder(b []byte) FaultDecoder {
	return FaultDecoder{HeaderDecoder(b)}
}

// IsInvalid reports whether c is too short to contain the status field, has an
// unsupported version, or is not a fault.
func (c FaultDecoder) IsInvalid() bool {
	if c.HeaderDecoder.IsInvalid() {
		return true
	}
	if len(c.HeaderDecoder) < faultStatusEnd {
		return true
	}
	return c.PacketType() != RPC_TYPE_FAULT
}

// Status returns the status field, which is non-zero for a run-time fault and 0
// for an exception that the stub data specifies.
func (c FaultDecoder) Status() uint32 {
	return le.Uint32(c.HeaderDecoder[24:faultStatusEnd])
}

// An InvalidResponseError reports a malformed or truncated RPC response.
type InvalidResponseError struct {
	Message string
}

func (e *InvalidResponseError) Error() string {
	return e.Message
}

// maxResponseStub is the largest response stub that ReadResponse accepts.
const maxResponseStub = 16 << 20

// ReadResponse returns the stub of the RPC response with the given call ID.
// first holds the bytes of the response that the caller has already read, which
// may be empty or only the start of the first fragment; read reads further bytes
// from the pipe. A response too large for one fragment continues in further
// fragments, each a separate PDU, and its stub is the concatenation of the
// fragment stubs up to and including the fragment with RPC_PACKET_FLAG_LAST set.
//
// If the server returns a fault PDU, the error reports its status. For a
// malformed or truncated response, or a stub larger than 16 MiB, the error is
// an *InvalidResponseError. Errors from read are returned unchanged.
func ReadResponse(first []byte, callId uint32, read func([]byte) (int, error)) ([]byte, error) {
	var stub []byte
	next := first
	for {
		frag, err := readFragment(next, read)
		if err != nil {
			return nil, err
		}
		next = nil

		if HeaderDecoder(frag).PacketType() == RPC_TYPE_FAULT {
			fault := NewFaultDecoder(frag)
			if fault.IsInvalid() || fault.CallId() != callId {
				return nil, &InvalidResponseError{"broken rpc fault format"}
			}
			return nil, fmt.Errorf("rpc fault with status 0x%08x", fault.Status())
		}

		r := NewResponseDecoder(frag)
		if r.IsInvalid() || r.CallId() != callId {
			return nil, &InvalidResponseError{"broken rpc response format"}
		}

		if len(stub)+len(r.Stub()) > maxResponseStub {
			return nil, &InvalidResponseError{"rpc response exceeds 16 MiB"}
		}
		stub = append(stub, r.Stub()...)

		if r.PacketFlags()&RPC_PACKET_FLAG_LAST != 0 {
			return stub, nil
		}
	}
}

// readFragment returns the RPC fragment that begins with b, calling read for
// the rest of it. If b is empty, readFragment reads the next fragment.
func readFragment(b []byte, read func([]byte) (int, error)) ([]byte, error) {
	if len(b) == 0 {
		buf := make([]byte, MaxXmitFrag)
		n, err := read(buf)
		if err != nil {
			return nil, err
		}
		b = buf[:n]
	}

	if len(b) < CommonHeaderSize {
		return nil, &InvalidResponseError{"truncated rpc fragment"}
	}

	for fragLen := int(HeaderDecoder(b).FragLength()); len(b) < fragLen; {
		rest := make([]byte, fragLen-len(b))
		n, err := read(rest)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, &InvalidResponseError{"truncated rpc fragment"}
		}
		b = append(b, rest[:n]...)
	}

	return b, nil
}
