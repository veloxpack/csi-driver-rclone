package proton

import (
	"errors"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

type Calendar struct {
	ID          string
	Name        string
	Description string
	Color       string
	Display     Bool

	Type  CalendarType
	Flags CalendarFlag
}

type CalendarFlag int64

const (
	CalendarFlagActive CalendarFlag = 1 << iota
	CalendarFlagUpdatePassphrase
	CalendarFlagResetNeeded
	CalendarFlagIncompleteSetup
	CalendarFlagLostAccess
)

type CalendarType int

const (
	CalendarTypeNormal CalendarType = iota
	CalendarTypeSubscribed
)

type CalendarKey struct {
	ID           string
	CalendarID   string
	PassphraseID string
	PrivateKey   string
	Flags        CalendarKeyFlag
}

func (key CalendarKey) Unlock(passphrase []byte) (*crypto.Key, error) {
	lockedKey, err := crypto.NewKeyFromArmored(key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return lockedKey.Unlock(passphrase)
}

type CalendarKeys []CalendarKey

func (keys CalendarKeys) Unlock(passphrase []byte) (*crypto.KeyRing, error) {
	kr, err := crypto.NewKeyRing(nil)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if k, err := key.Unlock(passphrase); err != nil {
			continue
		} else if err := kr.AddKey(k); err != nil {
			return nil, err
		}
	}

	return kr, nil
}

// TODO: What is this?
type CalendarKeyFlag int64

const (
	CalendarKeyFlagActive CalendarKeyFlag = 1 << iota
	CalendarKeyFlagPrimary
)

type CalendarMember struct {
	ID          string
	Permissions CalendarPermissions
	Email       string
	Color       string
	Display     Bool
	CalendarID  string
}

// TODO: What is this?
type CalendarPermissions int

// TODO: Support invitations.
type CalendarPassphrase struct {
	ID                string
	Flags             CalendarPassphraseFlag
	MemberPassphrases []MemberPassphrase
}

func (passphrase CalendarPassphrase) Decrypt(memberID string, addrKR *crypto.KeyRing) ([]byte, error) {
	for _, passphrase := range passphrase.MemberPassphrases {
		if passphrase.MemberID == memberID {
			return passphrase.decrypt(addrKR)
		}
	}

	return nil, errors.New("no such member passphrase")
}

// TODO: What is this?
type CalendarPassphraseFlag int64

type MemberPassphrase struct {
	MemberID   string
	Passphrase string
	Signature  string
}

func (passphrase MemberPassphrase) decrypt(addrKR *crypto.KeyRing) ([]byte, error) {
	msg, err := crypto.NewPGPMessageFromArmored(passphrase.Passphrase)
	if err != nil {
		return nil, err
	}

	dec, err := decryptMessage(addrKR, msg, nil, 0)
	if err != nil {
		return nil, err
	}

	if err := verifyDetachedArmored(addrKR, dec, passphrase.Signature, 0); err != nil {
		return nil, err
	}

	return dec, nil
}
