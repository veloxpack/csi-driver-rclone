package proton

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

type LinkWalkFunc func([]string, Link, *crypto.KeyRing) error

// Link holds the tree structure, for the clients, they represent the files and folders of a given volume.
// They have a ParentLinkID that points to parent folders.
// Links also hold the file name (encrypted) and a hash of the name for name collisions.
// Link data is encrypted with its owning Share keyring.
type Link struct {
	LinkID       string // Encrypted file/folder ID
	ParentLinkID string // Encrypted parent folder ID (LinkID). Root link has null ParentLinkID.

	Type               LinkType
	Name               string // Encrypted file name
	NameSignatureEmail string // Signature email for link name
	Hash               string // HMAC of name encrypted with parent hash key
	Size               int64
	State              LinkState
	MIMEType           string

	CreateTime     int64 // Link creation time
	ModifyTime     int64 // Link modification time (on API, real modify date is stored in XAttr)
	ExpirationTime int64 // Link expiration time

	NodeKey                 string // The private NodeKey, used to decrypt any file/folder content.
	NodePassphrase          string // The passphrase used to unlock the NodeKey, encrypted by the owning Link/Share keyring.
	NodePassphraseSignature string
	SignatureEmail          string // Signature email for the NodePassphraseSignature
	XAttr                   string // Modification time and size from the file system

	FileProperties   *FileProperties
	FolderProperties *FolderProperties
}

type LinkState int

const (
	LinkStateDraft LinkState = iota
	LinkStateActive
	LinkStateTrashed
	LinkStateDeleted
	LinkStateRestoring
)

func (l Link) GetName(parentNodeKR, addrKR *crypto.KeyRing) (string, error) {
	encName, err := crypto.NewPGPMessageFromArmored(l.Name)
	if err != nil {
		return "", err
	}

	decName, err := decryptMessage(parentNodeKR, encName, addrKR, 0)
	if err != nil {
		return "", err
	}

	return string(decName), nil
}

func (l Link) GetKeyRing(parentNodeKR, addrKR *crypto.KeyRing) (*crypto.KeyRing, error) {
	enc, err := crypto.NewPGPMessageFromArmored(l.NodePassphrase)
	if err != nil {
		return nil, err
	}

	dec, err := decryptMessage(parentNodeKR, enc, nil, 0)
	if err != nil {
		return nil, err
	}

	if err := verifyDetachedArmored(addrKR, dec, l.NodePassphraseSignature, 0); err != nil {
		return nil, err
	}

	lockedKey, err := crypto.NewKeyFromArmored(l.NodeKey)
	if err != nil {
		return nil, err
	}

	unlockedKey, err := lockedKey.Unlock(dec)
	if err != nil {
		return nil, err
	}

	return crypto.NewKeyRing(unlockedKey)
}

func (l Link) GetHashKey(parentNodeKey, addrKRs *crypto.KeyRing) ([]byte, error) {
	if l.Type != LinkTypeFolder {
		return nil, errors.New("link is not a folder")
	}

	enc, err := crypto.NewPGPMessageFromArmored(l.FolderProperties.NodeHashKey)
	if err != nil {
		return nil, err
	}

	_, ok := enc.SignatureKeyIDs()
	if ok {
		return decryptMessage(parentNodeKey, enc, addrKRs, 0)
	}
	return decryptMessage(parentNodeKey, enc, nil, 0)
}

func (l Link) GetSessionKey(nodeKR *crypto.KeyRing) (*crypto.SessionKey, error) {
	if l.Type != LinkTypeFile {
		return nil, errors.New("link is not a file")
	}

	dec, err := base64.StdEncoding.DecodeString(l.FileProperties.ContentKeyPacket)
	if err != nil {
		return nil, err
	}

	key, err := decryptSessionKey(nodeKR, dec)
	if err != nil {
		return nil, err
	}

	if err := verifyDetachedArmored(nodeKR, key.Key, l.FileProperties.ContentKeyPacketSignature, 0); err != nil {
		return nil, err
	}

	return key, nil
}

type FileProperties struct {
	ContentKeyPacket          string           // The block's key packet, encrypted with the node key.
	ContentKeyPacketSignature string           // Signature of the content key packet. Signature of the session key, signed with the NodeKey.
	ActiveRevision            RevisionMetadata // The active revision of the file.
}

type FolderProperties struct {
	NodeHashKey string // HMAC key used to hash the folder's children names.
}

type LinkType int

const (
	LinkTypeFolder LinkType = iota + 1
	LinkTypeFile
)

type RevisionMetadata struct {
	ID                string        // Encrypted Revision ID
	CreateTime        int64         // Unix timestamp of the revision creation time
	Size              int64         // Size of the revision in bytes
	ManifestSignature string        // Signature of the revision manifest, signed with user's address key of the share.
	SignatureEmail    string        // Email of the user that signed the revision.
	State             RevisionState // State of revision
	XAttr             string        // modification time and size from the file system
	Thumbnail         Bool          // Whether the revision has a thumbnail
	ThumbnailHash     string        // Hash of the thumbnail
}

func (revisionMetadata *RevisionMetadata) GetDecXAttrString(addrKR, nodeKR *crypto.KeyRing) (*RevisionXAttrCommon, error) {
	if revisionMetadata.XAttr == "" {
		return nil, nil
	}

	// decrypt the modification time and size
	XAttrMsg, err := crypto.NewPGPMessageFromArmored(revisionMetadata.XAttr)
	if err != nil {
		return nil, err
	}

	decXAttr, err := decryptMessage(nodeKR, XAttrMsg, addrKR, 0)
	if err != nil {
		return nil, err
	}

	var data RevisionXAttr
	err = json.Unmarshal(decXAttr, &data)
	if err != nil {
		// TODO: if Unmarshal fails, maybe it's because the file system is missing the field?
		return nil, err
	}

	return &data.Common, nil
}

// Revisions are only for files, they represent “versions” of files.
// Each file can have 1 active revision and n obsolete revisions.
type Revision struct {
	RevisionMetadata

	Blocks []Block
}

func (revision *Revision) GetDecXAttrString(addrKR, nodeKR *crypto.KeyRing) (*RevisionXAttrCommon, error) {
	if revision.XAttr == "" {
		return nil, nil
	}

	// decrypt the modification time and size
	XAttrMsg, err := crypto.NewPGPMessageFromArmored(revision.XAttr)
	if err != nil {
		return nil, err
	}

	decXAttr, err := decryptMessage(nodeKR, XAttrMsg, addrKR, 0)
	if err != nil {
		return nil, err
	}

	var data RevisionXAttr
	err = json.Unmarshal(decXAttr, &data)
	if err != nil {
		// TODO: if Unmarshal fails, maybe it's because the file system is missing the field?
		return nil, err
	}

	return &data.Common, nil
}

type RevisionState int

const (
	RevisionStateDraft RevisionState = iota
	RevisionStateActive
	RevisionStateObsolete
	RevisionStateDeleted
)

type CheckAvailableHashesReq struct {
	Hashes []string
}

type PendingHashData struct {
	Hash       []string
	RevisionID []string
	LinkID     []string
}
type CheckAvailableHashesRes struct {
	AvailableHashes   []string
	PendingHashesData []PendingHashData
}
