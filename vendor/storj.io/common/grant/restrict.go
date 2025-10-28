// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package grant

import (
	"errors"
	"strings"
	"time"

	"storj.io/common/encryption"
	"storj.io/common/macaroon"
	"storj.io/common/paths"
)

// SharePrefix defines a prefix that will be shared.
type SharePrefix struct {
	Bucket string
	// Prefix is the prefix of the shared object keys.
	//
	// Note: that within a bucket, the hierarchical key derivation scheme is
	// delineated by forward slashes (/), so encryption information will be
	// included in the resulting access grant to decrypt any key that shares
	// the same prefix up until the last slash.
	Prefix string
}

// Permission defines what actions can be used to share.
type Permission struct {
	// AllowDownload gives permission to download the object's content. It
	// allows getting object metadata, but it does not allow listing buckets.
	AllowDownload bool
	// AllowUpload gives permission to create buckets and upload new objects.
	// It does not allow overwriting existing objects unless AllowDelete is
	// granted too.
	AllowUpload bool
	// AllowList gives permission to list buckets. It allows getting object
	// metadata, but it does not allow downloading the object's content.
	AllowList bool
	// AllowDelete gives permission to delete buckets and objects. Unless
	// either AllowDownload or AllowList is granted too, no object metadata and
	// no error info will be returned for deleted objects.
	AllowDelete bool
	// AllowLock gives permission for retention periods to be placed on and
	// retrieved from objects. It also gives permission for Object Lock
	// configurations to be placed on and retrieved from buckets.
	//
	// Deprecated: AllowLock exists for historical compatibility
	// and should not be used. Prefer using the granular Object Lock
	// permissions AllowPutObjectRetention and AllowGetObjectRetention.
	AllowLock bool
	// AllowPutObjectRetention gives permission for retention periods to be
	// placed on and retrieved from objects.
	AllowPutObjectRetention bool
	// AllowGetObjectRetention gives permission for retention periods to be
	// retrieved from objects.
	AllowGetObjectRetention bool
	// AllowPutObjectLegalHold gives permission for legal hold status to be
	// placed on objects.
	AllowPutObjectLegalHold bool
	// AllowGetObjectLegalHold gives permission for legal hold status to be
	// retrieved from objects.
	AllowGetObjectLegalHold bool
	// AllowBypassGovernanceRetention gives permission for governance retention
	// to be bypassed on objects.
	AllowBypassGovernanceRetention bool
	// AllowPutBucketObjectLockConfiguration gives permission for object lock configuration to be
	// placed on buckets.
	AllowPutBucketObjectLockConfiguration bool
	// AllowGetBucketObjectLockConfiguration gives permission for object lock configuration to be
	// retrieved from buckets.
	AllowGetBucketObjectLockConfiguration bool
	// NotBefore restricts when the resulting access grant is valid for.
	// If set, the resulting access grant will not work if the Satellite
	// believes the time is before NotBefore.
	// If set, this value should always be before NotAfter.
	NotBefore time.Time
	// NotAfter restricts when the resulting access grant is valid for.
	// If set, the resulting access grant will not work if the Satellite
	// believes the time is after NotAfter.
	// If set, this value should always be after NotBefore.
	NotAfter time.Time
	// MaxObjectTTL restricts the maximum time-to-live of objects.
	// If set, new objects are uploaded with an expiration time that reflects
	// the MaxObjectTTL period.
	// If objects are uploaded with an explicit expiration time, the upload
	// will be successful only if it is shorter than the MaxObjectTTL period.
	MaxObjectTTL *time.Duration
}

// Restrict creates a new access grant with specific permissions.
//
// Access grants can only have their existing permissions restricted,
// and the resulting access grant will only allow for the intersection of all previous
// Restrict calls in the access grant construction chain.
//
// Prefixes, if provided, restrict the access grant (and internal encryption information)
// to only contain enough information to allow access to just those prefixes.
func (access *Access) Restrict(permission Permission, prefixes ...SharePrefix) (*Access, error) {
	if permission == (Permission{}) {
		return nil, errors.New("permission is empty")
	}

	var notBefore, notAfter *time.Time
	if !permission.NotBefore.IsZero() {
		notBefore = &permission.NotBefore
	}
	if !permission.NotAfter.IsZero() {
		notAfter = &permission.NotAfter
	}

	if notBefore != nil && notAfter != nil && notAfter.Before(*notBefore) {
		return nil, errors.New("invalid time range")
	}

	if permission.MaxObjectTTL != nil && *(permission.MaxObjectTTL) <= 0 {
		return nil, errors.New("non-positive ttl period")
	}

	caveat := macaroon.WithNonce(macaroon.Caveat{
		DisallowReads:                            !permission.AllowDownload,
		DisallowWrites:                           !permission.AllowUpload,
		DisallowLists:                            !permission.AllowList,
		DisallowDeletes:                          !permission.AllowDelete,
		DisallowLocks:                            !permission.AllowLock,
		DisallowPutRetention:                     !permission.AllowPutObjectRetention,
		DisallowGetRetention:                     !permission.AllowGetObjectRetention,
		DisallowPutLegalHold:                     !permission.AllowPutObjectLegalHold,
		DisallowGetLegalHold:                     !permission.AllowGetObjectLegalHold,
		DisallowBypassGovernanceRetention:        !permission.AllowBypassGovernanceRetention,
		DisallowPutBucketObjectLockConfiguration: !permission.AllowPutBucketObjectLockConfiguration,
		DisallowGetBucketObjectLockConfiguration: !permission.AllowGetBucketObjectLockConfiguration,
		NotBefore:                                notBefore,
		NotAfter:                                 notAfter,
		MaxObjectTtl:                             permission.MaxObjectTTL,
	})

	for _, prefix := range prefixes {
		// If the share prefix ends in a `/` we need to remove this final slash.
		// Otherwise, if we the shared prefix is `/bob/`, the encrypted shared
		// prefix results in `enc("")/enc("bob")/enc("")`. This is an incorrect
		// encrypted prefix, what we really want is `enc("")/enc("bob")`.
		unencPath := paths.NewUnencrypted(strings.TrimSuffix(prefix.Prefix, "/"))

		encPath, err := encryption.EncryptPathWithStoreCipher(prefix.Bucket, unencPath, access.EncAccess.Store)
		if err != nil {
			return nil, err
		}

		caveat.AllowedPaths = append(caveat.AllowedPaths, &macaroon.Caveat_Path{
			Bucket:              []byte(prefix.Bucket),
			EncryptedPathPrefix: []byte(encPath.Raw()),
		})
	}

	restrictedAPIKey, err := access.APIKey.Restrict(caveat)
	if err != nil {
		return nil, err
	}

	encAccess := access.EncAccess.Clone()
	encAccess.LimitTo(restrictedAPIKey)

	return &Access{
		SatelliteAddress: access.SatelliteAddress,
		APIKey:           restrictedAPIKey,
		EncAccess:        encAccess,
	}, nil
}
