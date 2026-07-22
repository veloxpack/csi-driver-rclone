package smb2

import (
	"context"
	"fmt"

	"github.com/cloudsoda/go-smb2/internal/smb2"
	"github.com/cloudsoda/go-smb2/internal/utf16le"
)

// sentinelFileId is used for related compound operations.
// The server interprets it as "use the handle from the preceding CREATE".
var sentinelFileId = &smb2.FileId{
	Persistent: [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	Volatile:   [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
}

// compoundSecurityResult holds the result of one file's security query.
type compoundSecurityResult struct {
	data []byte // raw security descriptor bytes; nil on error
	err  error
}

// compoundSecurityInfoBatch queries security descriptors for multiple files
// using compound CREATE+QUERY_INFO+CLOSE requests. It sub-batches internally
// based on available credits and returns one result per path (in order).
func (tc *treeConn) compoundSecurityInfoBatch(
	paths []string,
	securityInfo uint32,
	mapping utf16le.MapChars,
	ctx context.Context,
) ([]compoundSecurityResult, error) {
	results := make([]compoundSecurityResult, len(paths))

	// Determine access rights.
	access := uint32(smb2.READ_CONTROL)
	if securityInfo&smb2.SACL_SECUIRTY_INFORMATION != 0 {
		access |= smb2.ACCESS_SYSTEM_SECURITY
	}

	for off := 0; off < len(paths); {
		remaining := len(paths) - off

		// Loan credits: 3 per file (CREATE + QUERY_INFO + CLOSE).
		// Compute in int to avoid uint16 overflow on large directories,
		// then clamp to the credit balance capacity.
		wanted := min(remaining*3, cap(tc.account.balance))

		granted, _, err := tc.account.loan(ctx, uint16(wanted))
		if err != nil {
			return nil, err
		}

		batchSize := int(granted / 3)
		if batchSize == 0 {
			tc.chargeCredit(granted)
			return nil, &InternalError{"insufficient credits for compound request"}
		}
		if batchSize > remaining {
			batchSize = remaining
		}

		// Return excess credits.
		if excess := granted - uint16(batchSize*3); excess > 0 {
			tc.chargeCredit(excess)
		}

		err = tc.sendSecurityBatch(paths[off:off+batchSize], results[off:off+batchSize], access, securityInfo, mapping, ctx)
		if err != nil {
			return nil, err
		}

		off += batchSize
	}

	return results, nil
}

// sendSecurityBatch sends one compound batch and populates results.
func (tc *treeConn) sendSecurityBatch(
	paths []string,
	results []compoundSecurityResult,
	access, securityInfo uint32,
	mapping utf16le.MapChars,
	ctx context.Context,
) error {
	n := len(paths)
	entries := make([]compoundEntry, n*3)

	for i, path := range paths {
		base := i * 3

		// CREATE — first in each related triplet.
		entries[base] = compoundEntry{
			req: &smb2.CreateRequest{
				PacketHeader:         smb2.PacketHeader{CreditCharge: 1},
				RequestedOplockLevel: smb2.SMB2_OPLOCK_LEVEL_NONE,
				ImpersonationLevel:   smb2.Impersonation,
				DesiredAccess:        access,
				FileAttributes:       smb2.FILE_ATTRIBUTE_NORMAL,
				ShareAccess:          smb2.FILE_SHARE_READ,
				CreateDisposition:    smb2.FILE_OPEN,
				Name:                 path,
				Mapping:              mapping,
			},
			tc: tc,
		}

		// QUERY_INFO — related, uses sentinel FileId.
		entries[base+1] = compoundEntry{
			req: &smb2.QueryInfoRequest{
				PacketHeader:          smb2.PacketHeader{CreditCharge: 1},
				InfoType:              smb2.SMB2_0_INFO_SECURITY,
				FileInfoClass:         0,
				OutputBufferLength:    64 * 1024,
				AdditionalInformation: securityInfo,
				FileId:                sentinelFileId,
			},
			tc:      tc,
			related: true,
		}

		// CLOSE — related, uses sentinel FileId.
		entries[base+2] = compoundEntry{
			req: &smb2.CloseRequest{
				PacketHeader: smb2.PacketHeader{CreditCharge: 1},
				FileId:       sentinelFileId,
			},
			tc:      tc,
			related: true,
		}
	}

	rrs, err := tc.sendCompound(ctx, entries)
	if err != nil {
		return err
	}

	// Receive all responses. Each triplet: CREATE, QUERY_INFO, CLOSE.
	for i := range paths {
		base := i * 3

		// CREATE response.
		createPkt, createErr := tc.recv(rrs[base])
		if createErr != nil {
			results[i].err = createErr
			// Still drain QUERY_INFO and CLOSE responses.
			tc.recv(rrs[base+1])
			tc.recv(rrs[base+2])
			continue
		}
		if _, createErr = accept(smb2.SMB2_CREATE, createPkt); createErr != nil {
			results[i].err = createErr
			tc.recv(rrs[base+1])
			tc.recv(rrs[base+2])
			continue
		}

		// QUERY_INFO response — extract security descriptor.
		qiPkt, qiErr := tc.recv(rrs[base+1])
		if qiErr != nil {
			results[i].err = qiErr
			tc.recv(rrs[base+2])
			continue
		}
		qiRes, qiErr := accept(smb2.SMB2_QUERY_INFO, qiPkt)
		if qiErr != nil {
			results[i].err = qiErr
			tc.recv(rrs[base+2])
			continue
		}

		r := smb2.QueryInfoResponseDecoder(qiRes)
		if r.IsInvalid() {
			results[i].err = &InvalidResponseError{"broken query info response format"}
			tc.recv(rrs[base+2])
			continue
		}
		results[i].data = r.OutputBuffer()

		// CLOSE response — just drain it.
		tc.recv(rrs[base+2])
	}

	return nil
}

type treeConn struct {
	*session
	treeId     uint32
	shareFlags uint32

	// path string
	// shareType  uint8
	// capabilities uint32
	// maximalAccess uint32
}

func treeConnect(ctx context.Context, s *session, path string, flags uint16, mc utf16le.MapChars) (*treeConn, error) {
	req := &smb2.TreeConnectRequest{
		Flags:   flags,
		Path:    path,
		Mapping: mc,
	}

	req.CreditCharge = 1

	rr, err := s.send(ctx, req)
	if err != nil {
		return nil, err
	}

	pkt, err := s.recv(rr)
	if err != nil {
		return nil, err
	}

	res, err := accept(smb2.SMB2_TREE_CONNECT, pkt)
	if err != nil {
		return nil, err
	}

	r := smb2.TreeConnectResponseDecoder(res)
	if r.IsInvalid() {
		return nil, &InvalidResponseError{"broken tree connect response format"}
	}

	tc := &treeConn{
		session:    s,
		treeId:     smb2.PacketCodec(pkt).TreeId(),
		shareFlags: r.ShareFlags(),
		// path:    path,
		// shareType:  r.ShareType(),
		// capabilities: r.Capabilities(),
		// maximalAccess: r.MaximalAccess(),
	}

	return tc, nil
}

func (tc *treeConn) disconnect(ctx context.Context) error {
	req := new(smb2.TreeDisconnectRequest)

	req.CreditCharge = 1

	res, err := tc.sendRecv(ctx, smb2.SMB2_TREE_DISCONNECT, req)
	if err != nil {
		return err
	}

	r := smb2.TreeDisconnectResponseDecoder(res)
	if r.IsInvalid() {
		return &InvalidResponseError{"broken tree disconnect response format"}
	}

	return nil
}

func (tc *treeConn) sendRecv(ctx context.Context, cmd uint16, req smb2.Packet) (res []byte, err error) {
	rr, err := tc.send(ctx, req)
	if err != nil {
		return nil, err
	}

	pkt, err := tc.recv(rr)
	if err != nil {
		return nil, err
	}

	return accept(cmd, pkt)
}

func (tc *treeConn) send(ctx context.Context, req smb2.Packet) (rr *requestResponse, err error) {
	return tc.sendWith(ctx, req, tc)
}

func (tc *treeConn) recv(rr *requestResponse) (pkt []byte, err error) {
	pkt, err = tc.session.recv(rr)
	if err != nil {
		return nil, err
	}
	if rr.asyncId != 0 {
		if asyncId := smb2.PacketCodec(pkt).AsyncId(); asyncId != rr.asyncId {
			return nil, &InvalidResponseError{fmt.Sprintf("expected async id: %v, got %v", rr.asyncId, asyncId)}
		}
	} else {
		if treeId := smb2.PacketCodec(pkt).TreeId(); treeId != tc.treeId {
			return nil, &InvalidResponseError{fmt.Sprintf("expected tree id: %v, got %v", tc.treeId, treeId)}
		}
	}
	return pkt, err
}
