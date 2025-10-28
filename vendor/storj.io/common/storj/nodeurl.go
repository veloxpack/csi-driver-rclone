// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/zeebo/errs"

	"storj.io/common/base58"
)

var (
	// ErrNodeURL is used when something goes wrong with a node url.
	ErrNodeURL = errs.Class("node URL")
)

// NodeURL defines a structure for connecting to a node.
type NodeURL struct {
	ID            NodeID
	Address       string
	NoiseInfo     NoiseInfo
	DebounceLimit int
	Features      uint64 // this is a bitmask of pb.NodeAddress_Feature values.
}

// ParseNodeURL parses node URL string.
//
// Examples:
//
//	raw IP:
//	  33.20.0.1:7777
//	  [2001:db8:1f70::999:de8:7648:6e8]:7777
//
//	with NodeID:
//	  12vha9oTFnerxYRgeQ2BZqoFrLrnmmf5UWTCY2jA77dF3YvWew7@33.20.0.1:7777
//	  12vha9oTFnerxYRgeQ2BZqoFrLrnmmf5UWTCY2jA77dF3YvWew7@[2001:db8:1f70::999:de8:7648:6e8]:7777
//
//	without host:
//	  12vha9oTFnerxYRgeQ2BZqoFrLrnmmf5UWTCY2jA77dF3YvWew7@
//
//	with noise information:
//	  12vha9oTFnerxYRgeQ2BZqoFrLrnmmf5UWTCY2jA77dF3YvWew7@33.20.0.1:7777?noise_pub=12vha9oTFnerxY&noise_proto=1
func ParseNodeURL(s string) (NodeURL, error) {
	if s == "" {
		return NodeURL{}, nil
	}
	if !strings.HasPrefix(s, "storj://") {
		if !strings.Contains(s, "://") {
			s = "storj://" + s
		}
	}

	u, err := url.Parse(s)
	if err != nil {
		return NodeURL{}, ErrNodeURL.Wrap(err)
	}
	if u.Scheme != "" && u.Scheme != "storj" {
		return NodeURL{}, ErrNodeURL.New("unknown scheme %q", u.Scheme)
	}

	var node NodeURL
	if u.User != nil {
		node.ID, err = NodeIDFromString(u.User.String())
		if err != nil {
			return NodeURL{}, ErrNodeURL.Wrap(err)
		}
	}
	node.Address = u.Host

	query := u.Query()
	if query.Get("noise_pub") != "" {
		pubKey, _, err := base58.CheckDecode(query.Get("noise_pub"))
		if err != nil {
			return NodeURL{}, ErrNodeURL.Wrap(err)
		}
		node.NoiseInfo.PublicKey = string(pubKey)
	}
	if query.Get("noise_proto") != "" {
		protoInt, err := strconv.Atoi(query.Get("noise_proto"))
		if err != nil {
			return NodeURL{}, ErrNodeURL.Wrap(err)
		}
		node.NoiseInfo.Proto = NoiseProto(protoInt)
	}
	if query.Get("debounce") != "" {
		debounceInt, err := strconv.Atoi(query.Get("debounce"))
		if err != nil {
			return NodeURL{}, ErrNodeURL.Wrap(err)
		}
		node.DebounceLimit = debounceInt
	}
	if query.Get("f") != "" {
		features, err := strconv.ParseUint(query.Get("f"), 16, 64)
		if err != nil {
			return NodeURL{}, ErrNodeURL.Wrap(err)
		}
		node.Features = features
	}

	return node, nil
}

// IsZero returns whether the url is empty.
func (u NodeURL) IsZero() bool {
	return u == NodeURL{}
}

// String converts NodeURL to a string.
func (u NodeURL) String() string {
	hasSuffix := u.DebounceLimit > 0 || u.Features > 0 || !u.NoiseInfo.IsZero()
	if !hasSuffix {
		if u.ID.IsZero() {
			return u.Address
		}
		return u.ID.String() + "@" + u.Address
	}

	var s strings.Builder
	s.Grow(160)
	if !u.ID.IsZero() {
		s.WriteString(u.ID.String())
		s.WriteString("@")
	}
	s.WriteString(u.Address)

	const maxIntBytes = 22
	writeInt := func(v int64, base int) {
		var buf [maxIntBytes]byte
		s.Write(strconv.AppendInt(buf[:0], v, base))
	}
	writeUint := func(v uint64, base int) {
		var buf [maxIntBytes]byte
		s.Write(strconv.AppendUint(buf[:0], v, base))
	}

	delim := byte('?')
	writeKey := func(key string) {
		s.WriteByte(delim)
		delim = '&'
		s.WriteString(key)
	}

	if u.DebounceLimit > 0 {
		writeKey("debounce=")
		writeInt(int64(u.DebounceLimit), 10)
	}
	if u.Features > 0 {
		writeKey("f=")
		writeUint(u.Features, 16)
	}

	info := &u.NoiseInfo
	if info.Proto != NoiseProto_Unset {
		writeKey("noise_proto=")
		writeInt(int64(info.Proto), 10)
	}
	if info.PublicKey != "" {
		writeKey("noise_pub=")
		// base58 is URL safe
		s.WriteString(base58.CheckEncode([]byte(info.PublicKey), 0))
	}

	return s.String()
}

// Set implements flag.Value interface.
func (u *NodeURL) Set(s string) error {
	parsed, err := ParseNodeURL(s)
	if err != nil {
		return ErrNodeURL.Wrap(err)
	}

	*u = parsed
	return nil
}

// Type implements pflag.Value.
func (NodeURL) Type() string { return "storj.NodeURL" }

// NodeURLs defines a comma delimited flag for defining a list node url-s.
type NodeURLs []NodeURL

// ParseNodeURLs parses comma delimited list of node urls.
func ParseNodeURLs(s string) (NodeURLs, error) {
	var urls NodeURLs
	if s == "" {
		return nil, nil
	}

	for _, s := range strings.Split(s, ",") {
		u, err := ParseNodeURL(s)
		if err != nil {
			return nil, ErrNodeURL.Wrap(err)
		}
		urls = append(urls, u)
	}

	return urls, nil
}

// String converts NodeURLs to a string.
func (urls NodeURLs) String() string {
	var xs []string
	for _, u := range urls {
		xs = append(xs, u.String())
	}
	return strings.Join(xs, ",")
}

// Set implements flag.Value interface.
func (urls *NodeURLs) Set(s string) error {
	parsed, err := ParseNodeURLs(s)
	if err != nil {
		return ErrNodeURL.Wrap(err)
	}

	*urls = parsed
	return nil
}

// Type implements pflag.Value.
func (NodeURLs) Type() string { return "storj.NodeURLs" }
