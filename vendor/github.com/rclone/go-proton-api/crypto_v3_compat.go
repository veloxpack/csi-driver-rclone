package proton

// This file provides small helpers that mirror the v2 KeyRing.* API surface
// on top of the v3 gopenpgp builder API. They keep call sites compact during
// (and after) the v2 → v3 migration.
//
// Note: these helpers deliberately do NOT call ClearPrivateParams() on the
// returned v3 handles. In v3 the handle holds pointers into the caller's
// keyring, so clearing the handle would also zero the keyring's private key
// material and break any subsequent use of the same keyring (which is the
// common case here — the same KeyRing is used across many operations).

import (
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/armor"
	"github.com/ProtonMail/gopenpgp/v3/constants"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// serverTimeUnix stores the most recently observed server Unix time, sampled
// from the Date response header. v2 used a global crypto.UpdateTime for this
// purpose; v3 has no global clock, so we track it ourselves and pass it into
// per-operation builders. A zero value means "not set yet".
var serverTimeUnix atomic.Int64

func setServerTime(unixTime int64) {
	serverTimeUnix.Store(unixTime)
}

// resolveVerifyTime returns the verification time to pass to v3 builders.
// A non-zero caller-supplied verifyTime is used as-is; otherwise the most
// recently sampled server time is used, falling back to local time. This
// mirrors v2's crypto.GetUnixTime() behaviour.
func resolveVerifyTime(verifyTime int64) int64 {
	if verifyTime != 0 {
		return verifyTime
	}
	if t := serverTimeUnix.Load(); t != 0 {
		return t
	}
	return time.Now().Unix()
}

// encryptMessage encrypts plaintext bytes to the given recipients, optionally
// signing with signKR. Equivalent to v2 kr.Encrypt(NewPlainMessage(b), signKR).
func encryptMessage(kr *crypto.KeyRing, plaintext []byte, signKR *crypto.KeyRing) (*crypto.PGPMessage, error) {
	builder := crypto.PGP().Encryption().Recipients(kr)
	if signKR != nil {
		builder = builder.SigningKeys(signKR)
	}
	eh, err := builder.New()
	if err != nil {
		return nil, err
	}
	return eh.Encrypt(plaintext)
}

// EncryptMessageNonAead encrypts plaintext bytes to the given recipients,
// optionally signing with signKR, always producing the pre-crypto-refresh
// wire format: a v3 PKESK followed by a v1 SEIPD packet.
//
// Encrypting via the recipients path (encryptMessage) selects the
// crypto-refresh format (v6 PKESK + v2 SEIPD with AEAD) whenever every
// recipient key advertises support for it, which Proton Drive's v6 file node
// keys do. Proton Drive requires the old format for everything except file
// content — names, node passphrases, extended attributes and block signatures
// must not use AEAD regardless of the recipient key's preferences, otherwise
// the official Proton clients fail to decrypt them. To sidestep the format
// negotiation, the data is encrypted with an explicitly generated non-v6
// session key (yielding the v1 SEIPD) and the session key is wrapped for the
// recipients separately (yielding the v3 PKESK).
func EncryptMessageNonAead(kr *crypto.KeyRing, plaintext []byte, signKR *crypto.KeyRing) (*crypto.PGPMessage, error) {
	sk, err := crypto.GenerateSessionKeyAlgo(constants.AES256)
	if err != nil {
		return nil, err
	}
	keyPacket, err := encryptSessionKey(kr, sk)
	if err != nil {
		return nil, err
	}
	builder := crypto.PGP().Encryption().SessionKey(sk)
	if signKR != nil {
		builder = builder.SigningKeys(signKR)
	}
	eh, err := builder.New()
	if err != nil {
		return nil, err
	}
	dataMsg, err := eh.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return crypto.NewPGPMessage(append(keyPacket, dataMsg.Bytes()...)), nil
}

// decryptMessage decrypts a PGP message, optionally verifying signatures
// against verifyKR using verifyTime (or the sampled server / local time if
// verifyTime is zero). Equivalent to v2 kr.Decrypt(msg, verifyKR, t).
func decryptMessage(kr *crypto.KeyRing, msg *crypto.PGPMessage, verifyKR *crypto.KeyRing, verifyTime int64) ([]byte, error) {
	builder := crypto.PGP().Decryption().DecryptionKeys(kr)
	if verifyKR != nil {
		builder = builder.VerificationKeys(verifyKR).VerifyTime(resolveVerifyTime(verifyTime))
	}
	dh, err := builder.New()
	if err != nil {
		return nil, err
	}

	result, err := dh.Decrypt(msg.Bytes(), crypto.Bytes)
	if err != nil {
		return nil, err
	}
	if verifyKR != nil {
		if sigErr := result.SignatureError(); sigErr != nil {
			return nil, sigErr
		}
	}
	return result.Bytes(), nil
}

// signDetached produces a binary detached signature for the plaintext bytes.
// Equivalent to v2 kr.SignDetached(NewPlainMessage(b)).Bytes().
func signDetached(kr *crypto.KeyRing, plaintext []byte) ([]byte, error) {
	sh, err := crypto.PGP().Sign().SigningKeys(kr).Detached().New()
	if err != nil {
		return nil, err
	}
	return sh.Sign(plaintext, crypto.Bytes)
}

// signDetachedArmored produces an armored detached signature for the
// plaintext bytes. Equivalent to v2 kr.SignDetached(NewPlainMessage(b)).GetArmored().
func signDetachedArmored(kr *crypto.KeyRing, plaintext []byte) (string, error) {
	sh, err := crypto.PGP().Sign().SigningKeys(kr).Detached().New()
	if err != nil {
		return "", err
	}
	sig, err := sh.Sign(plaintext, crypto.Armor)
	if err != nil {
		return "", err
	}
	return string(sig), nil
}

// verifyDetached verifies a binary detached signature over plaintext.
// If verifyTime is zero the sampled server / local time is used (matching
// v2 crypto.GetUnixTime() semantics). Equivalent to v2
// kr.VerifyDetached(NewPlainMessage(b), sig, t).
func verifyDetached(kr *crypto.KeyRing, plaintext, signature []byte, verifyTime int64) error {
	vh, err := crypto.PGP().Verify().VerificationKeys(kr).VerifyTime(resolveVerifyTime(verifyTime)).New()
	if err != nil {
		return err
	}
	result, err := vh.VerifyDetached(plaintext, signature, crypto.Bytes)
	if err != nil {
		return err
	}
	return result.SignatureError()
}

// verifyDetachedArmored verifies an armored detached signature over plaintext.
func verifyDetachedArmored(kr *crypto.KeyRing, plaintext []byte, armoredSig string, verifyTime int64) error {
	vh, err := crypto.PGP().Verify().VerificationKeys(kr).VerifyTime(resolveVerifyTime(verifyTime)).New()
	if err != nil {
		return err
	}
	result, err := vh.VerifyDetached(plaintext, []byte(armoredSig), crypto.Armor)
	if err != nil {
		return err
	}
	return result.SignatureError()
}

// encryptAttachment encrypts attachment bytes, returning the PGP message
// (which can be split into key + data packets via BinaryKeyPacket /
// BinaryDataPacket). The filename argument is ignored — v3 does not expose
// a way to set the literal-data filename through the public API.
func encryptAttachment(kr *crypto.KeyRing, plaintext []byte, _ string) (*crypto.PGPMessage, error) {
	eh, err := crypto.PGP().Encryption().Recipients(kr).New()
	if err != nil {
		return nil, err
	}
	return eh.Encrypt(plaintext)
}

// decryptSessionKey extracts the session key from a key-packet using the
// decryption keyring. Equivalent to v2 kr.DecryptSessionKey(keyPacket).
func decryptSessionKey(kr *crypto.KeyRing, keyPacket []byte) (*crypto.SessionKey, error) {
	dh, err := crypto.PGP().Decryption().DecryptionKeys(kr).New()
	if err != nil {
		return nil, err
	}
	return dh.DecryptSessionKey(keyPacket)
}

// encryptSessionKey wraps a session key into a key-packet for the given
// recipient keyring. Equivalent to v2 kr.EncryptSessionKey(sk).
func encryptSessionKey(kr *crypto.KeyRing, sk *crypto.SessionKey) ([]byte, error) {
	eh, err := crypto.PGP().Encryption().Recipients(kr).New()
	if err != nil {
		return nil, err
	}
	return eh.EncryptSessionKey(sk)
}

// armorSignature wraps a binary signature in PGP armor.
func armorSignature(sig []byte) (string, error) {
	return armor.ArmorPGPSignature(sig)
}

// unarmorSignature unwraps an armored signature into binary.
func unarmorSignature(armored string) ([]byte, error) {
	return armor.Unarmor(armored)
}
