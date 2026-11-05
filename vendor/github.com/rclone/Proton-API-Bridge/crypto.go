package proton_api_bridge

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/ProtonMail/gopenpgp/v3/armor"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	proton "github.com/rclone/go-proton-api"
)

// protonDrivePGP returns a gopenpgp handle configured for Proton Drive's
// crypto-refresh (RFC 9580) file-content format: v6 keys/key packets (PKESK)
// and a v2 SEIPD data packet using AES-256-GCM. The RFC9580 profile defaults
// the AEAD mode to OCB, so we override it to GCM (the mode Proton Drive
// requires).
//
// It is used to generate v6 file node keys (see generateLockedKey) and to
// encrypt file blocks. When encrypting blocks, the content session key's v6
// flag selects the v2 SEIPD packet and this handle's AEAD config selects the
// GCM mode; a non-v6 session key (an older file's existing content key) yields
// a v1 SEIPD instead, keeping revisions consistent with the original format.
func protonDrivePGP() *crypto.PGPHandle {
	p := profile.RFC9580()
	p.AeadEncryption = &packet.AEADConfig{DefaultMode: packet.AEADModeGCM}
	return crypto.PGPWithProfile(p)
}

func generatePassphrase() (string, error) {
	token, err := crypto.RandomToken(32)
	if err != nil {
		return "", err
	}

	tokenBase64 := base64.StdEncoding.EncodeToString(token)
	return tokenBase64, nil
}

// generateLockedKey is the v3-equivalent of v2 helper.GenerateKey: it generates
// a new key with the given identity, locks it with the passphrase, and returns
// the armored locked key.
//
// When aead is true the key is generated with the crypto-refresh handle so it
// is a v6 key. Node keys are currently always generated with aead false (v4):
// the official Proton clients cannot yet decrypt files whose node key is v6,
// so originating crypto-refresh files is not safe. If v6 file node keys are
// ever enabled again, note that Proton's server rejects a v6 content key
// packet (PKESK) encrypted to a v4 node key, so the two must change together.
func generateLockedKey(name, email string, passphrase []byte, aead bool) (string, error) {
	pgp := crypto.PGP()
	if aead {
		pgp = protonDrivePGP()
	}
	key, err := pgp.KeyGeneration().AddUserId(name, email).New().GenerateKey()
	if err != nil {
		return "", err
	}
	defer key.ClearPrivateParams()

	locked, err := pgp.LockKey(key, passphrase)
	if err != nil {
		return "", err
	}
	return locked.Armor()
}

func generateCryptoKey(aead bool) (string, string, error) {
	passphrase, err := generatePassphrase()
	if err != nil {
		return "", "", err
	}

	// "Drive key" / "noreply@protonmail.com" are the hardcoded identity used by
	// the iOS Drive client; v3's default profile produces a curve25519 key,
	// which is what the v2 helper produced for keyType "x25519".
	key, err := generateLockedKey("Drive key", "noreply@protonmail.com", []byte(passphrase), aead)
	if err != nil {
		return "", "", err
	}

	return passphrase, key, nil
}

// encryptWithSignature encrypts b to kr and signs it with addrKR, returning the
// armored ciphertext and armored detached signature.
//
// Node passphrases must stay in the pre-crypto-refresh format regardless of
// the recipient key's preferences (see proton.EncryptMessageNonAead).
func encryptWithSignature(kr, addrKR *crypto.KeyRing, b []byte) (string, string, error) {
	pgp := crypto.PGP()

	enc, err := proton.EncryptMessageNonAead(kr, b, nil)
	if err != nil {
		return "", "", err
	}

	encArm, err := enc.Armor()
	if err != nil {
		return "", "", err
	}

	signHandle, err := pgp.Sign().SigningKeys(addrKR).Detached().New()
	if err != nil {
		return "", "", err
	}

	sigArm, err := signHandle.Sign(b, crypto.Armor)
	if err != nil {
		return "", "", err
	}

	return encArm, string(sigArm), nil
}

// generateNodeKeys generates a node key, its passphrase, and the passphrase
// signature. aead selects a v6 (crypto-refresh) node key (currently unused —
// see generateLockedKey). The passphrase is always encrypted to the parent
// keyring in the pre-crypto-refresh format.
func generateNodeKeys(kr, addrKR *crypto.KeyRing, aead bool) (string, string, string, error) {
	nodePassphrase, nodeKey, err := generateCryptoKey(aead)
	if err != nil {
		return "", "", "", err
	}

	nodePassphraseEnc, nodePassphraseSignature, err := encryptWithSignature(kr, addrKR, []byte(nodePassphrase))
	if err != nil {
		return "", "", "", err
	}

	return nodeKey, nodePassphraseEnc, nodePassphraseSignature, nil
}

func reencryptKeyPacket(srcKR, dstKR, addrKR *crypto.KeyRing, passphrase string) (string, error) {
	pgp := crypto.PGP()

	oldMessage, err := crypto.NewPGPMessageFromArmored(passphrase)
	if err != nil {
		return "", err
	}

	srcDec, err := pgp.Decryption().DecryptionKeys(srcKR).New()
	if err != nil {
		return "", err
	}

	sessionKey, err := srcDec.DecryptSessionKey(oldMessage.BinaryKeyPacket())
	if err != nil {
		return "", err
	}

	dstEnc, err := pgp.Encryption().Recipients(dstKR).New()
	if err != nil {
		return "", err
	}

	newKeyPacket, err := dstEnc.EncryptSessionKey(sessionKey)
	if err != nil {
		return "", err
	}

	newSplitMessage := crypto.NewPGPSplitMessage(newKeyPacket, oldMessage.BinaryDataPacket())
	return newSplitMessage.Armor()
}

func getKeyRing(kr, addrKR *crypto.KeyRing, key, passphrase, passphraseSignature string) (*crypto.KeyRing, error) {
	pgp := crypto.PGP()

	enc, err := crypto.NewPGPMessageFromArmored(passphrase)
	if err != nil {
		return nil, err
	}

	decHandle, err := pgp.Decryption().DecryptionKeys(kr).New()
	if err != nil {
		return nil, err
	}

	decResult, err := decHandle.Decrypt(enc.Bytes(), crypto.Bytes)
	if err != nil {
		return nil, err
	}
	dec := decResult.Bytes()

	sig, err := armor.Unarmor(passphraseSignature)
	if err != nil {
		return nil, err
	}

	verifyHandle, err := pgp.Verify().VerificationKeys(addrKR).New()
	if err != nil {
		return nil, err
	}
	verifyResult, err := verifyHandle.VerifyDetached(dec, sig, crypto.Bytes)
	if err != nil {
		return nil, err
	}
	if sigErr := verifyResult.SignatureError(); sigErr != nil {
		return nil, sigErr
	}

	lockedKey, err := crypto.NewKeyFromArmored(key)
	if err != nil {
		return nil, err
	}

	unlockedKey, err := lockedKey.Unlock(dec)
	if err != nil {
		return nil, err
	}

	return crypto.NewKeyRing(unlockedKey)
}

func decryptBlockIntoBuffer(sessionKey *crypto.SessionKey, addrKR, nodeKR *crypto.KeyRing, originalHash, encSignature string, buffer io.ReaderFrom, block io.ReadCloser) error {
	data, err := io.ReadAll(block)
	if err != nil {
		return err
	}

	skDec, err := crypto.PGP().Decryption().SessionKey(sessionKey).New()
	if err != nil {
		return err
	}
	skResult, err := skDec.Decrypt(data, crypto.Bytes)
	if err != nil {
		return err
	}
	plain := skResult.Bytes()

	encSignatureMsg, err := crypto.NewPGPMessageFromArmored(encSignature)
	if err != nil {
		return err
	}

	// v2 used addrKR.VerifyDetachedEncrypted(plainMessage, encSignature, nodeKR, time),
	// which decrypts the encrypted detached signature with nodeKR and then verifies it
	// against the already-decrypted plain data using addrKR. v3 has no single-call
	// equivalent: DecryptDetached treats its first argument as encrypted PGP data and
	// would feed our plaintext into openpgp.ReadMessage. So we do it in two steps:
	// decrypt the signature with nodeKR, then verify it against the plain data with addrKR.
	sigDecHandle, err := crypto.PGP().Decryption().DecryptionKeys(nodeKR).New()
	if err != nil {
		return err
	}
	sigDecResult, err := sigDecHandle.Decrypt(encSignatureMsg.Bytes(), crypto.Bytes)
	if err != nil {
		return err
	}
	sigBytes := sigDecResult.Bytes()

	verifyHandle, err := crypto.PGP().Verify().VerificationKeys(addrKR).New()
	if err != nil {
		return err
	}
	verifyResult, err := verifyHandle.VerifyDetached(plain, sigBytes, crypto.Bytes)
	if err != nil {
		return err
	}
	if sigErr := verifyResult.SignatureError(); sigErr != nil {
		return sigErr
	}

	if _, err := buffer.ReadFrom(bytes.NewReader(plain)); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)
	base64Hash := base64.StdEncoding.EncodeToString(hash)
	if base64Hash != originalHash {
		return ErrDownloadedBlockHashVerificationFailed
	}

	return nil
}
