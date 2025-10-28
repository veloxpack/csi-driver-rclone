package smb2

import (
	"encoding/asn1"
	"errors"
	"fmt"
	"time"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/crypto"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/iana/flags"
	"github.com/jcmturner/gokrb5/v8/iana/keyusage"
	"github.com/jcmturner/gokrb5/v8/messages"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/jcmturner/gokrb5/v8/types"
)

// Krb5Initiator is a GSSAPI initiator for Kerberos 5.
// It implements the Initiator interface.
type Krb5Initiator struct {
	Client    *client.Client
	TargetSPN string

	sessKey    types.EncryptionKey
	sessSubkey types.EncryptionKey
}

// OID returns the Kerberos 5 OID.
func (ki *Krb5Initiator) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier(gssapi.OIDKRB5.OID())
}

// InitSecContext initiates the security context.
func (ki *Krb5Initiator) InitSecContext() ([]byte, error) {
	if ki.Client == nil {
		return nil, errors.New("Kerberos client is not set")
	}

	if ki.TargetSPN == "" {
		return nil, errors.New("Kerberos target SPN is not set")
	}

	tkt, key, err := ki.Client.GetServiceTicket(ki.TargetSPN)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kerberos service ticket: %w", err)
	}

	req, err := spnego.NewKRB5TokenAPREQ(ki.Client, tkt, key, []int{gssapi.ContextFlagMutual, gssapi.ContextFlagConf}, []int{flags.APOptionMutualRequired})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kerberos AP-REQ: %w", err)
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Kerberos AP-REQ: %w", err)
	}

	ki.sessKey = key

	return reqBytes, nil
}

// AcceptSecContext accepts the security context.
func (ki *Krb5Initiator) AcceptSecContext(sc []byte) ([]byte, error) {
	var token spnego.KRB5Token
	err := token.Unmarshal(sc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kerberos token: %w", err)
	}

	if !token.IsAPRep() {
		return nil, fmt.Errorf("expected Kerberos AP-REP, got: %#v", token)
	}

	data, err := crypto.DecryptEncPart(token.APRep.EncPart, ki.sessKey, keyusage.AP_REP_ENCPART)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt Kerberos AP-REP EncPart: %w", err)
	}

	var payload messages.EncAPRepPart
	if err := payload.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Kerberos AP-REP EncPart: %w", err)
	}

	if time.Since(payload.CTime).Abs() > ki.Client.Config.LibDefaults.Clockskew {
		return nil, fmt.Errorf("AP_REP time out of range (now: %v, AP_REP time: %v)", time.Now().UTC().Truncate(time.Second), payload.CTime)
	}

	ki.sessSubkey = payload.Subkey

	return []byte{}, nil
}

// Sum creates a checksum with the session subkey.
func (ki *Krb5Initiator) Sum(bs []byte) []byte {
	token, err := gssapi.NewInitiatorMICToken(bs, ki.sessSubkey)
	if err != nil {
		return nil
	}

	b, err := token.Marshal()
	if err != nil {
		return nil
	}

	return b
}

// SessionKey returns the session key.
func (ki *Krb5Initiator) SessionKey() []byte {
	// Only the first 16 bytes are used, if less than that are available
	// zero padding is added.
	sliced := ki.sessSubkey.KeyValue[:16]

	for len(sliced) < 16 {
		sliced = append(sliced, 0x00)
	}

	return sliced
}
