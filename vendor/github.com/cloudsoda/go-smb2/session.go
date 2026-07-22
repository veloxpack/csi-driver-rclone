package smb2

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"

	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"

	"github.com/cloudsoda/go-smb2/internal/crypto/ccm"
	"github.com/cloudsoda/go-smb2/internal/crypto/cmac"
	"github.com/cloudsoda/go-smb2/internal/erref"
	"github.com/cloudsoda/go-smb2/internal/smb2"
)

func sessionSetup(ctx context.Context, conn *conn, i Initiator) (*session, error) {
	spnego := newSpnegoClient([]Initiator{i})

	outputToken, err := spnego.InitSecContext()
	if err != nil {
		return nil, &InvalidResponseError{err.Error()}
	}

	req := &smb2.SessionSetupRequest{
		Flags:             0,
		Capabilities:      conn.capabilities & (smb2.SMB2_GLOBAL_CAP_DFS),
		Channel:           0,
		SecurityBuffer:    outputToken,
		PreviousSessionId: 0,
	}

	if conn.requireSigning {
		req.SecurityMode = smb2.SMB2_NEGOTIATE_SIGNING_REQUIRED
	} else {
		req.SecurityMode = smb2.SMB2_NEGOTIATE_SIGNING_ENABLED
	}

	req.CreditCharge = 1
	req.CreditRequestResponse = conn.account.initRequest()

	rr, err := conn.send(ctx, req)
	if err != nil {
		return nil, err
	}

	pkt, err := conn.recv(rr)
	if err != nil {
		return nil, err
	}

	p := smb2.PacketCodec(pkt)

	status := erref.NtStatus(p.Status())
	if status != erref.STATUS_MORE_PROCESSING_REQUIRED && status != erref.STATUS_SUCCESS {
		return nil, &InvalidResponseError{fmt.Sprintf("expected status: %v or %v, got %v", erref.STATUS_MORE_PROCESSING_REQUIRED, erref.STATUS_SUCCESS, erref.NtStatus(p.Status()))}
	}

	res, err := accept(smb2.SMB2_SESSION_SETUP, pkt)
	if err != nil {
		return nil, err
	}

	r := smb2.SessionSetupResponseDecoder(res)
	if r.IsInvalid() {
		return nil, &InvalidResponseError{"broken session setup response format"}
	}

	sessionFlags := r.SessionFlags()
	if conn.requireSigning {
		if sessionFlags&smb2.SMB2_SESSION_FLAG_IS_GUEST != 0 {
			return nil, &InvalidResponseError{"guest account doesn't support signing"}
		}
		if sessionFlags&smb2.SMB2_SESSION_FLAG_IS_NULL != 0 {
			return nil, &InvalidResponseError{"anonymous account doesn't support signing"}
		}
	}

	s := &session{
		conn:         conn,
		sessionFlags: sessionFlags,
		sessionId:    p.SessionId(),
	}

	switch conn.dialect {
	case smb2.SMB311:
		s.preauthIntegrityHashValue = conn.preauthIntegrityHashValue

		switch conn.preauthIntegrityHashId {
		case smb2.SHA512:
			h := sha512.New()
			h.Write(s.preauthIntegrityHashValue[:])
			h.Write(rr.pkt)
			h.Sum(s.preauthIntegrityHashValue[:0])

			if status == erref.STATUS_MORE_PROCESSING_REQUIRED {
				h.Reset()
				h.Write(s.preauthIntegrityHashValue[:])
				h.Write(pkt)
				h.Sum(s.preauthIntegrityHashValue[:0])
			}
		}

	}

	outputToken, err = spnego.AcceptSecContext(r.SecurityBuffer())
	if err != nil {
		return nil, &InvalidResponseError{err.Error()}
	}

	// We set session before sending packet just for setting hdr.SessionId.
	// But, we should not permit access from receiver until the session information is completed.
	conn.session.Store(s)

	if status == erref.STATUS_MORE_PROCESSING_REQUIRED {
		req.SecurityBuffer = outputToken

		req.CreditRequestResponse = 0

		rr, err = s.send(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	if s.sessionFlags&(smb2.SMB2_SESSION_FLAG_IS_GUEST|smb2.SMB2_SESSION_FLAG_IS_NULL) == 0 {
		sessionKey := spnego.SessionKey()

		switch conn.dialect {
		case smb2.SMB202, smb2.SMB210:
			s.signer = hmac.New(sha256.New, sessionKey)
			s.verifier = hmac.New(sha256.New, sessionKey)
		case smb2.SMB300, smb2.SMB302:
			signingKey := kdf(sessionKey, []byte("SMB2AESCMAC\x00"), []byte("SmbSign\x00"))
			ciph, err := aes.NewCipher(signingKey)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}
			s.signer = cmac.New(ciph)
			s.verifier = cmac.New(ciph)

			// s.applicationKey = kdf(sessionKey, []byte("SMB2APP\x00"), []byte("SmbRpc\x00"))

			encryptionKey := kdf(sessionKey, []byte("SMB2AESCCM\x00"), []byte("ServerIn \x00"))
			decryptionKey := kdf(sessionKey, []byte("SMB2AESCCM\x00"), []byte("ServerOut\x00"))

			ciph, err = aes.NewCipher(encryptionKey)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}
			s.encrypter, err = ccm.NewCCMWithNonceAndTagSizes(ciph, 11, 16)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}

			ciph, err = aes.NewCipher(decryptionKey)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}
			s.decrypter, err = ccm.NewCCMWithNonceAndTagSizes(ciph, 11, 16)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}
		case smb2.SMB311:
			if status == erref.STATUS_MORE_PROCESSING_REQUIRED {
				switch conn.preauthIntegrityHashId {
				case smb2.SHA512:
					h := sha512.New()
					h.Write(s.preauthIntegrityHashValue[:])
					h.Write(rr.pkt)
					h.Sum(s.preauthIntegrityHashValue[:0])
				}
			}

			signingKey := kdf(sessionKey, []byte("SMBSigningKey\x00"), s.preauthIntegrityHashValue[:])
			ciph, err := aes.NewCipher(signingKey)
			if err != nil {
				return nil, &InternalError{err.Error()}
			}
			s.signer = cmac.New(ciph)
			s.verifier = cmac.New(ciph)

			// s.applicationKey = kdf(sessionKey, []byte("SMBAppKey\x00"), preauthIntegrityHashValue)

			encryptionKey := kdf(sessionKey, []byte("SMBC2SCipherKey\x00"), s.preauthIntegrityHashValue[:])
			decryptionKey := kdf(sessionKey, []byte("SMBS2CCipherKey\x00"), s.preauthIntegrityHashValue[:])

			switch s.cipherId {
			case smb2.AES128CCM:
				ciph, err := aes.NewCipher(encryptionKey)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
				s.encrypter, err = ccm.NewCCMWithNonceAndTagSizes(ciph, 11, 16)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}

				ciph, err = aes.NewCipher(decryptionKey)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
				s.decrypter, err = ccm.NewCCMWithNonceAndTagSizes(ciph, 11, 16)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
			case smb2.AES128GCM:
				ciph, err := aes.NewCipher(encryptionKey)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
				s.encrypter, err = cipher.NewGCMWithNonceSize(ciph, 12)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}

				ciph, err = aes.NewCipher(decryptionKey)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
				s.decrypter, err = cipher.NewGCMWithNonceSize(ciph, 12)
				if err != nil {
					return nil, &InternalError{err.Error()}
				}
			}
		}
	}

	if status == erref.STATUS_MORE_PROCESSING_REQUIRED {
		pkt, err = s.recv(rr)
		if err != nil {
			return nil, err
		}

		res, err = accept(smb2.SMB2_SESSION_SETUP, pkt)
		if err != nil {
			return nil, err
		}

		r = smb2.SessionSetupResponseDecoder(res)
		if r.IsInvalid() {
			return nil, &InvalidResponseError{"broken session setup response format"}
		}

		if erref.NtStatus(smb2.PacketCodec(pkt).Status()) != erref.STATUS_SUCCESS {
			return nil, &InvalidResponseError{"broken session setup response format"}
		}
	}

	s.sessionFlags = r.SessionFlags()

	// now, allow access from receiver
	s.useSession.Store(true)

	return s, nil
}

type session struct {
	*conn
	sessionFlags              uint16
	sessionId                 uint64
	preauthIntegrityHashValue [64]byte

	signer    hash.Hash
	verifier  hash.Hash
	encrypter cipher.AEAD
	decrypter cipher.AEAD

	encryptBuf []byte // reusable ciphertext buffer; safe because conn.m serializes access

	// applicationKey []byte
}

func (s *session) logoff(ctx context.Context) error {
	req := new(smb2.LogoffRequest)

	req.CreditCharge = 1

	_, err := s.sendRecv(ctx, smb2.SMB2_LOGOFF, req)
	if err != nil {
		return err
	}

	s.conn.rdone <- struct{}{}
	s.conn.t.Close()

	return nil
}

func (s *session) echo(ctx context.Context) error {
	req := new(smb2.EchoRequest)

	req.CreditCharge = 1

	res, err := s.sendRecv(ctx, smb2.SMB2_ECHO, req)
	if err != nil {
		return err
	}

	r := smb2.EchoResponseDecoder(res)
	if r.IsInvalid() {
		return &InvalidResponseError{"broken echo response format"}
	}

	return nil
}

func (s *session) sendRecv(ctx context.Context, cmd uint16, req smb2.Packet) (res []byte, err error) {
	rr, err := s.send(ctx, req)
	if err != nil {
		return nil, err
	}

	pkt, err := s.recv(rr)
	if err != nil {
		return nil, err
	}

	return accept(cmd, pkt)
}

func (s *session) recv(rr *requestResponse) (pkt []byte, err error) {
	pkt, err = s.conn.recv(rr)
	if err != nil {
		return nil, err
	}
	// IBM i NetServer (iSeries/AS400) assigns the session ID only in the
	// STATUS_MORE_PROCESSING_REQUIRED response, while the client's sessionId
	// is still 0. Adopt the server's session ID in that case.
	sessionId := smb2.PacketCodec(pkt).SessionId()
	if s.sessionId == 0 {
		s.sessionId = sessionId
	} else if sessionId != s.sessionId {
		return nil, &InvalidResponseError{fmt.Sprintf("expected session id: %v, got %v", s.sessionId, sessionId)}
	}
	return pkt, err
}

func (s *session) sign(pkt []byte) []byte {
	p := smb2.PacketCodec(pkt)

	p.SetFlags(p.Flags() | smb2.SMB2_FLAGS_SIGNED)

	h := s.signer

	h.Reset()

	h.Write(pkt)

	p.SetSignature(h.Sum(nil))

	return pkt
}

func (s *session) verify(pkt []byte) (ok bool) {
	p := smb2.PacketCodec(pkt)

	signature := append([]byte{}, p.Signature()...)

	p.SetSignature(zero[:])

	h := s.verifier

	h.Reset()

	h.Write(pkt)

	p.SetSignature(h.Sum(nil))

	return bytes.Equal(signature, p.Signature())
}

// encrypt encrypts pkt into c. len(c) must equal 52+len(pkt)+16.
// Returns the wire-ready packet (c re-sliced to exclude the trailing tag).
func (s *session) encrypt(pkt, c []byte) ([]byte, error) {
	t := smb2.TransformCodec(c)

	t.SetProtocolId()
	if err := t.GenerateNonce(s.encrypter.NonceSize()); err != nil {
		return nil, err
	}
	t.SetOriginalMessageSize(uint32(len(pkt)))
	t.SetFlags(smb2.Encrypted)
	t.SetSessionId(s.sessionId)

	s.encrypter.Seal(c[:52], t.Nonce(s.encrypter.NonceSize()), pkt, t.AssociatedData())

	t.SetSignature(c[len(c)-16:])

	c = c[:len(c)-16]

	return c, nil
}

// decrypt decrypts an SMB3 transform packet. c must be at least
// len(EncryptedData)+len(Signature) bytes; decrypt copies the
// ciphertext and tag into c and decrypts in-place.
func (s *session) decrypt(pkt, c []byte) ([]byte, error) {
	t := smb2.TransformCodec(pkt)

	c = c[:0]
	c = append(c, t.EncryptedData()...)
	c = append(c, t.Signature()...)

	return s.decrypter.Open(
		c[:0],
		t.Nonce(s.decrypter.NonceSize()),
		c,
		t.AssociatedData(),
	)
}
