// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package macaroon

import (
	"bytes"
	"context"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"

	"storj.io/common/base58"
	"storj.io/picobuf"
)

// revoker is supplied when checking a macaroon for validation.
type revoker interface {
	// Check is intended to return a bool if any of the supplied tails are revoked.
	Check(ctx context.Context, tails [][]byte) (bool, error)
}

var (
	// Error is a general API Key error.
	Error = errs.Class("api key")
	// ErrFormat means that the structural formatting of the API Key is invalid.
	ErrFormat = errs.Class("api key format")
	// ErrInvalid means that the API Key is improperly signed.
	ErrInvalid = errs.Class("api key invalid")
	// ErrUnauthorized means that the API key does not grant the requested permission.
	ErrUnauthorized = errs.Class("api key unauthorized")
	// ErrRevoked means the API key has been revoked.
	ErrRevoked = errs.Class("api key revocation")

	mon = monkit.Package()
)

// ActionType specifies the operation type being performed that the Macaroon will validate.
type ActionType int

const (
	// not using iota because these values are persisted in macaroons.
	_ ActionType = 0

	// ActionRead specifies a read operation.
	ActionRead ActionType = 1
	// ActionWrite specifies a write operation.
	ActionWrite ActionType = 2
	// ActionList specifies a list operation.
	ActionList ActionType = 3
	// ActionDelete specifies a delete operation.
	ActionDelete ActionType = 4
	// ActionProjectInfo requests project-level information.
	ActionProjectInfo ActionType = 5
	// ActionLock represents the following actions:
	//
	//   - Placement or retrieval of retention periods for an object
	//   - Placement or retrieval of Object Lock configurations for a bucket
	//
	// Deprecated: ActionLock exists for historical compatibility
	// and should not be used. Prefer using the granular Object Lock actions
	// ActionPutObjectRetention and ActionGetObjectRetention.
	ActionLock ActionType = 6
	// ActionPutObjectRetention specifies an action related to updating
	// Object Retention configuration.
	ActionPutObjectRetention ActionType = 7
	// ActionGetObjectRetention specifies an action related to retrieving
	// Object Retention configuration.
	ActionGetObjectRetention ActionType = 8
	// ActionPutObjectLegalHold specifies an action related to updating
	// Object Legal Hold configuration.
	ActionPutObjectLegalHold ActionType = 9
	// ActionGetObjectLegalHold specifies an action related to retrieving
	// Object Legal Hold configuration.
	ActionGetObjectLegalHold ActionType = 10
	// ActionBypassGovernanceRetention specifies an action related to bypassing
	// Object Governance Retention.
	ActionBypassGovernanceRetention ActionType = 11
	// ActionPutBucketObjectLockConfiguration specifies an action related to updating
	// Bucket Object Lock configuration.
	ActionPutBucketObjectLockConfiguration ActionType = 12
	// ActionGetBucketObjectLockConfiguration specifies an action related to retrieving
	// Bucket Object Lock configuration.
	ActionGetBucketObjectLockConfiguration ActionType = 13
)

// APIKeyVersion specifies the version of an API key.
type APIKeyVersion uint

const (
	// APIKeyVersionMin is the minimum API key version.
	// It is for API keys that only support read, write, list, delete,
	// and project info retrieval actions.
	APIKeyVersionMin APIKeyVersion = 0

	// APIKeyVersionObjectLock is the API key version that introduces support
	// for Object Lock actions.
	APIKeyVersionObjectLock APIKeyVersion = 1 << 0 // 0b001

	// APIKeyVersionAuditable is the API key version that introduces support
	// for auditability.
	APIKeyVersionAuditable APIKeyVersion = 1 << 1 // 0b010
)

// SupportsObjectLock returns true if the API key version supports Object Lock actions.
func (v APIKeyVersion) SupportsObjectLock() bool {
	return v&APIKeyVersionObjectLock != 0
}

// SupportsAuditability returns true if the API key is auditable.
func (v APIKeyVersion) SupportsAuditability() bool {
	return v&APIKeyVersionAuditable != 0
}

// Action specifies the specific operation being performed that the Macaroon will validate.
type Action struct {
	Op            ActionType
	Bucket        []byte
	EncryptedPath []byte
	Time          time.Time
}

// APIKey implements a Macaroon-backed Storj-v3 API key.
type APIKey struct {
	mac *Macaroon
}

// ParseAPIKey parses a given api key string and returns an APIKey if the
// APIKey was correctly formatted. It does not validate the key.
func ParseAPIKey(key string) (*APIKey, error) {
	data, version, err := base58.CheckDecode(key)
	if err != nil || version != 0 {
		return nil, ErrFormat.New("invalid api key format")
	}
	mac, err := ParseMacaroon(data)
	if err != nil {
		return nil, ErrFormat.Wrap(err)
	}
	return &APIKey{mac: mac}, nil
}

// ParseRawAPIKey parses raw api key data and returns an APIKey if the APIKey
// was correctly formatted. It does not validate the key.
func ParseRawAPIKey(data []byte) (*APIKey, error) {
	mac, err := ParseMacaroon(data)
	if err != nil {
		return nil, ErrFormat.Wrap(err)
	}
	return &APIKey{mac: mac}, nil
}

// NewAPIKey generates a brand new unrestricted API key given the provided.
// server project secret.
func NewAPIKey(secret []byte) (*APIKey, error) {
	mac, err := NewUnrestricted(secret)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return &APIKey{mac: mac}, nil
}

// FromParts generates an api key from the provided parts.
func FromParts(head, secret []byte, caveats ...Caveat) (_ *APIKey, err error) {
	apiKey := &APIKey{mac: NewUnrestrictedFromParts(head, secret)}

	for _, caveat := range caveats {
		apiKey, err = apiKey.Restrict(caveat)
		if err != nil {
			return nil, Error.Wrap(err)
		}
	}

	return apiKey, nil
}

// Check makes sure that the key authorizes the provided action given the root
// project secret, the API key's version, and any possible revocations, returning an error
// if the action is not authorized. 'revoked' is a list of revoked heads.
func (a *APIKey) Check(ctx context.Context, secret []byte, version APIKeyVersion, action Action, revoker revoker) (err error) {
	defer mon.Task()(&ctx)(&err)

	ok, tails := a.mac.ValidateAndTails(secret)
	if !ok {
		return ErrInvalid.New("macaroon unauthorized")
	}

	// a timestamp is always required on an action
	if action.Time.IsZero() {
		return Error.New("no timestamp provided")
	}

	if !version.SupportsObjectLock() {
		// API keys created before the introduction of granular Object Lock permissions
		// should be denied the ability to perform granular Object Lock actions.
		switch action.Op {
		case ActionPutObjectRetention,
			ActionGetObjectRetention,
			ActionPutObjectLegalHold,
			ActionGetObjectLegalHold,
			ActionBypassGovernanceRetention,
			ActionPutBucketObjectLockConfiguration,
			ActionGetBucketObjectLockConfiguration,
			ActionLock:
			return ErrUnauthorized.New("action disallowed")
		}
	}

	caveats := a.mac.Caveats()
	for _, cavbuf := range caveats {
		var cav Caveat
		if err := cav.UnmarshalBinary(cavbuf); err != nil {
			return ErrFormat.New("invalid caveat format")
		}
		if !cav.Allows(action) {
			return ErrUnauthorized.New("action disallowed")
		}
	}

	if revoker != nil {
		revoked, err := revoker.Check(ctx, tails)
		if err != nil {
			return ErrRevoked.Wrap(err)
		}
		if revoked {
			return ErrRevoked.New("contains revoked tail")
		}
	}

	return nil
}

// AllowedBuckets stores information about which buckets are
// allowed to be accessed, where `Buckets` stores names of buckets that are
// allowed and `All` is a bool that indicates if all buckets are allowed or not.
type AllowedBuckets struct {
	All     bool
	Buckets map[string]struct{}
}

// GetAllowedBuckets returns a list of all the allowed bucket paths that match the Action operation.
func (a *APIKey) GetAllowedBuckets(ctx context.Context, action Action) (allowed AllowedBuckets, err error) {
	defer mon.Task()(&ctx)(&err)

	// Every bucket is allowed until we find a caveat that restricts some paths.
	allowed.All = true

	// every caveat that includes a list of allowed paths must include the bucket for
	// the bucket to be allowed. in other words, the set of allowed buckets is the
	// intersection of all of the buckets in the allowed paths.
	for _, cavbuf := range a.mac.Caveats() {
		var cav Caveat
		if err := cav.UnmarshalBinary(cavbuf); err != nil {
			return AllowedBuckets{}, ErrFormat.New("invalid caveat format: %v", err)
		}
		if !cav.Allows(action) {
			return AllowedBuckets{}, ErrUnauthorized.New("action disallowed")
		}

		// If the caveat does not include any allowed paths, then it is not restricting it.
		if len(cav.AllowedPaths) == 0 {
			continue
		}

		// Since we found some path restrictions, it's definitely the case that not every
		// bucket is allowed.
		allowed.All = false

		caveatBuckets := map[string]struct{}{}
		for _, caveatPath := range cav.AllowedPaths {
			caveatBuckets[string(caveatPath.Bucket)] = struct{}{}
		}

		if allowed.Buckets == nil {
			allowed.Buckets = caveatBuckets
		} else {
			for bucket := range allowed.Buckets {
				if _, ok := caveatBuckets[bucket]; !ok {
					delete(allowed.Buckets, bucket)
				}
			}
		}
	}

	return allowed, err
}

// GetMaxObjectTTL returns the shortest MaxObjectTTL period conifgured in the APIKey's caveats.
func (a *APIKey) GetMaxObjectTTL(ctx context.Context) (ttl *time.Duration, err error) {
	defer mon.Task()(&ctx)(&err)

	caveats := a.mac.Caveats()
	for _, cavbuf := range caveats {
		var cav Caveat
		if err := cav.UnmarshalBinary(cavbuf); err != nil {
			return nil, ErrFormat.New("invalid caveat format")
		}
		if cav.MaxObjectTtl != nil && (ttl == nil || *(cav.MaxObjectTtl) < *ttl) {
			ttl = cav.MaxObjectTtl
		}
	}

	return ttl, nil
}

// Restrict generates a new APIKey with the provided Caveat attached.
func (a *APIKey) Restrict(caveat Caveat) (*APIKey, error) {
	buf, err := picobuf.Marshal(&caveat)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	mac, err := a.mac.AddFirstPartyCaveat(buf)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return &APIKey{mac: mac}, nil
}

// Head returns the identifier for this macaroon's root ancestor.
func (a *APIKey) Head() []byte {
	return a.mac.Head()
}

// Tail returns the identifier for this macaroon only.
func (a *APIKey) Tail() []byte {
	return a.mac.Tail()
}

// Serialize serializes the API Key to a string.
func (a *APIKey) Serialize() string {
	return base58.CheckEncode(a.mac.Serialize(), 0)
}

// SerializeRaw serialize the API Key to raw bytes.
func (a *APIKey) SerializeRaw() []byte {
	return a.mac.Serialize()
}

// Allows returns true if the provided action is allowed by the caveat.
func (c *Caveat) Allows(action Action) bool {
	// if the action is after the caveat's "not after" field, then it is invalid
	if c.NotAfter != nil && action.Time.After(*c.NotAfter) {
		return false
	}
	// if the caveat's "not before" field is *after* the action, then the action
	// is before the "not before" field and it is invalid
	if c.NotBefore != nil && c.NotBefore.After(action.Time) {
		return false
	}

	// we want to always allow reads for bucket metadata, perhaps filtered by the
	// buckets in the allowed paths.
	if action.Op == ActionRead && len(action.EncryptedPath) == 0 {
		if len(c.AllowedPaths) == 0 {
			return true
		}
		if len(action.Bucket) == 0 {
			// if no action.bucket name is provided, then this call is checking that
			// we can list all buckets. In that case, return true here and we will
			// filter out buckets that aren't allowed later with `GetAllowedBuckets()`
			return true
		}
		for _, path := range c.AllowedPaths {
			if bytes.Equal(path.Bucket, action.Bucket) {
				return true
			}
		}
		return false
	}

	switch action.Op {
	case ActionRead:
		if c.DisallowReads {
			return false
		}
	case ActionWrite:
		if c.DisallowWrites {
			return false
		}
	case ActionList:
		if c.DisallowLists {
			return false
		}
	case ActionDelete:
		if c.DisallowDeletes {
			return false
		}
	case ActionProjectInfo:
		// allow
	case ActionLock:
		if c.DisallowLocks {
			return false
		}
	case ActionPutObjectRetention:
		if c.DisallowPutRetention {
			return false
		}
	case ActionGetObjectRetention:
		if c.DisallowPutRetention {
			if c.DisallowGetRetention {
				return false
			}
		}
	case ActionPutObjectLegalHold:
		if c.DisallowPutLegalHold {
			return false
		}
	case ActionGetObjectLegalHold:
		if c.DisallowGetLegalHold {
			return false
		}
	case ActionBypassGovernanceRetention:
		if c.DisallowBypassGovernanceRetention {
			return false
		}
	case ActionPutBucketObjectLockConfiguration:
		if c.DisallowPutBucketObjectLockConfiguration {
			return false
		}
	case ActionGetBucketObjectLockConfiguration:
		if c.DisallowGetBucketObjectLockConfiguration {
			return false
		}
	default:
		return false
	}

	if len(c.AllowedPaths) > 0 && action.Op != ActionProjectInfo {
		found := false
		for _, path := range c.AllowedPaths {
			if bytes.Equal(action.Bucket, path.Bucket) &&
				bytes.HasPrefix(action.EncryptedPath, path.EncryptedPathPrefix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
