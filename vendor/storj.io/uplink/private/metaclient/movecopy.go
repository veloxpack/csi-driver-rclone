// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package metaclient

import (
	"context"
	"crypto/rand"

	"github.com/zeebo/errs"

	"storj.io/common/encryption"
	"storj.io/common/paths"
	"storj.io/common/pb"
	"storj.io/common/storj"
)

// EncryptedKeyAndNonce holds single segment encrypted key.
type EncryptedKeyAndNonce struct {
	Position          SegmentPosition
	EncryptedKeyNonce storj.Nonce
	EncryptedKey      []byte
}

// CopyObjectOptions options for CopyObject method.
type CopyObjectOptions struct {
	Retention Retention
}

// CopyObject atomically copies object to a different bucket or/and key. Source object version can be specified.
func (db *DB) CopyObject(ctx context.Context, sourceBucket, sourceKey string, sourceVersion []byte, targetBucket, targetKey string, opts CopyObjectOptions) (_ *Object, err error) {
	defer mon.Task()(&ctx)(&err)

	err = validateMoveCopyInput(sourceBucket, sourceKey, targetBucket, targetKey)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	sourceEncKey, err := encryption.EncryptPathWithStoreCipher(sourceBucket, paths.NewUnencrypted(sourceKey), db.encStore)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	targetEncKey, err := encryption.EncryptPathWithStoreCipher(targetBucket, paths.NewUnencrypted(targetKey), db.encStore)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	response, err := db.metainfo.BeginCopyObject(ctx, BeginCopyObjectParams{
		Bucket:                []byte(sourceBucket),
		EncryptedObjectKey:    []byte(sourceEncKey.Raw()),
		Version:               sourceVersion,
		NewBucket:             []byte(targetBucket),
		NewEncryptedObjectKey: []byte(targetEncKey.Raw()),
	})
	if err != nil {
		return nil, errs.Wrap(err)
	}

	oldDerivedKey, err := encryption.DeriveContentKey(sourceBucket, paths.NewUnencrypted(sourceKey), db.encStore)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	newDerivedKey, err := encryption.DeriveContentKey(targetBucket, paths.NewUnencrypted(targetKey), db.encStore)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	newMetadataEncryptedKey, newMetadataKeyNonce, err := db.reencryptMetadataKey(response.EncryptedMetadataKey, response.EncryptedMetadataKeyNonce, oldDerivedKey, newDerivedKey)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	newKeys, err := db.reencryptKeys(response.SegmentKeys, oldDerivedKey, newDerivedKey)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	params := FinishCopyObjectParams{
		StreamID:                     response.StreamID,
		NewBucket:                    []byte(targetBucket),
		NewEncryptedObjectKey:        []byte(targetEncKey.Raw()),
		NewEncryptedMetadataKeyNonce: newMetadataKeyNonce,
		NewEncryptedMetadataKey:      newMetadataEncryptedKey,
		NewSegmentKeys:               newKeys,
	}
	if opts != (CopyObjectOptions{}) {
		params.Retention = opts.Retention
	}
	obj, err := db.metainfo.FinishCopyObject(ctx, params)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	info, err := db.ObjectFromRawObjectItem(ctx, targetBucket, targetKey, obj.Info)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return &info, nil
}

func (db *DB) reencryptMetadataKey(encryptedMetadataKey []byte, encryptedMetadataKeyNonce storj.Nonce, oldDerivedKey, newDerivedKey *storj.Key) ([]byte, storj.Nonce, error) {
	if len(encryptedMetadataKey) == 0 {
		return nil, storj.Nonce{}, nil
	}

	cipherSuite := db.encryptionParameters.CipherSuite

	// decrypt old metadata key
	metadataContentKey, err := encryption.DecryptKey(encryptedMetadataKey, cipherSuite, oldDerivedKey, &encryptedMetadataKeyNonce)
	if err != nil {
		return nil, storj.Nonce{}, errs.Wrap(err)
	}

	// encrypt metadata content key with new derived key and old nonce
	newMetadataKeyNonce := encryptedMetadataKeyNonce
	newMetadataEncryptedKey, err := encryption.EncryptKey(metadataContentKey, cipherSuite, newDerivedKey, &newMetadataKeyNonce)
	if err != nil {
		return nil, storj.Nonce{}, errs.Wrap(err)
	}

	return newMetadataEncryptedKey, newMetadataKeyNonce, nil
}

func (db *DB) reencryptKeys(keys []EncryptedKeyAndNonce, oldDerivedKey, newDerivedKey *storj.Key) ([]EncryptedKeyAndNonce, error) {
	cipherSuite := db.encryptionParameters.CipherSuite

	newKeys := make([]EncryptedKeyAndNonce, len(keys))
	for i, oldKey := range keys {
		// decrypt old key
		contentKey, err := encryption.DecryptKey(oldKey.EncryptedKey, cipherSuite, oldDerivedKey, &oldKey.EncryptedKeyNonce)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		// create new random nonce and encrypt
		var newEncryptedKeyNonce storj.Nonce
		// generate random nonce for encrypting the content key
		_, err = rand.Read(newEncryptedKeyNonce[:])
		if err != nil {
			return nil, errs.Wrap(err)
		}

		newEncryptedKey, err := encryption.EncryptKey(contentKey, cipherSuite, newDerivedKey, &newEncryptedKeyNonce)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		newKeys[i] = EncryptedKeyAndNonce{
			Position:          oldKey.Position,
			EncryptedKeyNonce: newEncryptedKeyNonce,
			EncryptedKey:      newEncryptedKey,
		}
	}

	return newKeys, nil
}

func convertKeys(input []*pb.EncryptedKeyAndNonce) []EncryptedKeyAndNonce {
	keys := make([]EncryptedKeyAndNonce, len(input))
	for i, key := range input {
		keys[i] = EncryptedKeyAndNonce{
			EncryptedKeyNonce: key.EncryptedKeyNonce,
			EncryptedKey:      key.EncryptedKey,
		}
		if key.Position != nil {
			keys[i].Position = SegmentPosition{
				PartNumber: key.Position.PartNumber,
				Index:      key.Position.Index,
			}
		}
	}

	return keys
}

func validateMoveCopyInput(oldbucket, oldkey, newbucket, newkey string) error {
	switch {
	case oldbucket == "":
		return ErrNoBucket.New("%v", oldbucket)
	case oldkey == "":
		return ErrNoPath.New("%v", oldkey)
	case newbucket == "": // TODO should we make this error different
		return ErrNoBucket.New("%v", newbucket)
	case newkey == "": // TODO should we make this error different
		return ErrNoPath.New("%v", newkey)
	}

	return nil
}
