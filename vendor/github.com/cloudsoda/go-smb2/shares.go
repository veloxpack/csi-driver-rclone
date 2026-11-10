package smb2

import (
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/cloudsoda/go-smb2/internal/erref"
	"github.com/cloudsoda/go-smb2/internal/msrpc"
	"github.com/cloudsoda/go-smb2/internal/smb2"
	"github.com/cloudsoda/go-smb2/internal/srvsvc"
)

// ShareType is the kind of resource a share exposes. The values are the base
// share types defined in [MS-SRVS] section 2.2.2.4, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/6069f8c0-c93f-43a0-a5b4-7ed447eb4b84
type ShareType uint32

// Values of ShareType.
const (
	ShareTypeDiskTree   ShareType = srvsvc.STYPE_DISKTREE
	ShareTypePrintQueue ShareType = srvsvc.STYPE_PRINTQ
	ShareTypeDevice     ShareType = srvsvc.STYPE_DEVICE
	ShareTypeIPC        ShareType = srvsvc.STYPE_IPC
)

// String returns a lowercase name for t, or "unknown" if t is not one of the
// ShareType constants.
func (t ShareType) String() string {
	switch t {
	case ShareTypeDiskTree:
		return "disktree"
	case ShareTypePrintQueue:
		return "printq"
	case ShareTypeDevice:
		return "device"
	case ShareTypeIPC:
		return "ipc"
	default:
		return "unknown"
	}
}

// ShareInfo describes a share on a server, as returned by Session.ListShares.
// It contains the fields of a SHARE_INFO_1, defined in [MS-SRVS] section
// 2.2.4.23, at
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/fc69f110-998d-4c16-9667-514e22fdd80b
type ShareInfo struct {
	// Name is the share name, in the form accepted by Session.Mount.
	Name string

	// Comment is the share's description, as configured on the server.
	Comment string

	// TypeFlags is the share type as sent by the server: a base type in the
	// low bits and independent flags in the high bits. Type, IsSpecial and
	// IsTemporary decode it; TypeFlags also includes the bits they do not
	// decode.
	TypeFlags uint32
}

// Type returns the base type of the share, without the flag bits.
func (s ShareInfo) Type() ShareType {
	return ShareType(s.TypeFlags & srvsvc.STYPE_MASK)
}

// IsSpecial reports whether the share is marked as reserved for interprocess
// communication or remote administration. [MS-SRVS] gives IPC$ and ADMIN$ as
// examples, and states that the flag can also mark administrative shares such
// as C$.
func (s ShareInfo) IsSpecial() bool {
	return s.TypeFlags&srvsvc.STYPE_SPECIAL != 0
}

// IsTemporary reports whether the share is not persisted, so is not recreated
// when the file server next initializes.
func (s ShareInfo) IsTemporary() bool {
	return s.TypeFlags&srvsvc.STYPE_TEMPORARY != 0
}

// ListShares returns the shares exported by the server, including the type and
// the description the server reports for each one.
func (c *Session) ListShares() ([]ShareInfo, error) {
	shares, err := c.netShareEnum()
	if err != nil {
		return nil, err
	}

	infos := make([]ShareInfo, len(shares))
	for i, s := range shares {
		infos[i] = ShareInfo{
			Name:      s.Name,
			Comment:   s.Comment,
			TypeFlags: s.Type,
		}
	}

	return infos, nil
}

// ListSharenames returns the names of the shares exported by the server. Use
// ListShares to get the type and the description of each share as well.
func (c *Session) ListSharenames() ([]string, error) {
	shares, err := c.netShareEnum()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(shares))
	for i, s := range shares {
		names[i] = s.Name
	}

	return names, nil
}

// netShareEnum calls NetrShareEnum on the server service over the IPC$ share.
func (c *Session) netShareEnum() ([]srvsvc.ShareInfo1, error) {
	const op = "listShares"

	servername := c.addr
	if c.host != "" {
		servername = c.host
	}

	fs, err := c.Mount(fmt.Sprintf(`\\%s\IPC$`, servername))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fs.Umount()
	}()

	fs = fs.WithContext(c.ctx)

	// open the named pipe for the "server service"
	f, err := fs.OpenFile("srvsvc", os.O_RDWR, 0o666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	callId := rand.Uint32()

	bindReq := &smb2.IoctlRequest{
		CtlCode:           smb2.FSCTL_PIPE_TRANSCEIVE,
		OutputOffset:      0,
		OutputCount:       0,
		MaxInputResponse:  0,
		MaxOutputResponse: msrpc.MaxXmitFrag,
		Flags:             smb2.SMB2_0_IOCTL_IS_FSCTL,
		Input: &msrpc.Bind{
			CallId:                callId,
			InterfaceUUID:         srvsvc.UUID,
			InterfaceVersion:      srvsvc.VERSION,
			InterfaceVersionMinor: srvsvc.VERSION_MINOR,
		},
	}

	output, err := f.ioctl(bindReq)
	if err != nil {
		return nil, &os.PathError{Op: op, Path: f.name, Err: err}
	}

	r1 := msrpc.NewBindAckDecoder(output)
	if r1.IsInvalid() || r1.CallId() != callId {
		return nil, &os.PathError{Op: op, Path: f.name, Err: &InvalidResponseError{"broken bind ack response format"}}
	}
	if !r1.IsAccepted() {
		return nil, &os.PathError{Op: op, Path: f.name, Err: &InvalidResponseError{"server rejected the srvsvc bind"}}
	}

	callId++

	const level = srvsvc.ShareInfoLevel1

	reqReq := &smb2.IoctlRequest{
		CtlCode:          smb2.FSCTL_PIPE_TRANSCEIVE,
		OutputOffset:     0,
		OutputCount:      0,
		MaxInputResponse: 0,
		// This matches the max_recv_frag value sent in the bind, so a response
		// that fits in one fragment is returned in a single round trip. Larger
		// responses are reassembled below.
		MaxOutputResponse: msrpc.MaxXmitFrag,
		Flags:             smb2.SMB2_0_IOCTL_IS_FSCTL,
		Input: &msrpc.Request{
			CallId: callId,
			Opnum:  srvsvc.OP_NET_SHARE_ENUM,
			Stub: &srvsvc.NetShareEnumAllRequest{
				ServerName: servername,
				Level:      level,
			},
		},
	}

	output, err = f.ioctl(reqReq)
	if err != nil {
		rerr, ok := err.(*ResponseError)
		if !ok || erref.NtStatus(rerr.Code) != erref.STATUS_BUFFER_OVERFLOW {
			return nil, &os.PathError{Op: op, Path: f.name, Err: err}
		}
	}

	// output holds the start of the response, which is truncated if the ioctl
	// returned STATUS_BUFFER_OVERFLOW.
	stub, err := msrpc.ReadResponse(output, callId, func(b []byte) (int, error) {
		return f.readAt(b, 0)
	})
	if err != nil {
		var invalid *msrpc.InvalidResponseError
		if errors.As(err, &invalid) {
			err = &InvalidResponseError{invalid.Message}
		}
		return nil, &os.PathError{Op: op, Path: f.name, Err: err}
	}

	d := srvsvc.NetShareEnumAllResponseDecoder(stub)
	shares, rv, ok := d.Decode()
	if !ok {
		return nil, &os.PathError{Op: op, Path: f.name, Err: &InvalidResponseError{"broken net share enum response format"}}
	}
	if got, _ := d.Level(); got != level {
		return nil, &os.PathError{Op: op, Path: f.name, Err: &InvalidResponseError{fmt.Sprintf("net share enum response has level %d, requested %d", got, level)}}
	}
	if rv != 0 {
		return nil, &os.PathError{Op: op, Path: f.name, Err: fmt.Errorf("NetrShareEnum returned error 0x%08x", rv)}
	}

	return shares, nil
}
