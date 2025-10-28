package sddl

import (
	"encoding/binary"
	"fmt"
)

// FromBinary takes a binary security descriptor in relative format (contiguous memory with offsets)
func FromBinary(data []byte) (*SecurityDescriptor, error) {
	dataLen := uint32(len(data))
	if dataLen < 20 {
		return nil, fmt.Errorf("invalid security descriptor: it must be 20 bytes length at minimum")
	}

	revision := data[0]
	sbzl := data[1]
	control := binary.LittleEndian.Uint16(data[2:4])
	ownerOffset := binary.LittleEndian.Uint32(data[4:8])
	groupOffset := binary.LittleEndian.Uint32(data[8:12])
	saclOffset := binary.LittleEndian.Uint32(data[12:16])
	daclOffset := binary.LittleEndian.Uint32(data[16:20])

	if ownerOffset > 0 && ownerOffset >= dataLen {
		return nil, fmt.Errorf("invalid security descriptor: Owner offset 0x%x exceeds data length 0x%x", ownerOffset, dataLen)
	}
	if groupOffset > 0 && groupOffset >= dataLen {
		return nil, fmt.Errorf("invalid security descriptor: Group offset 0x%x exceeds data length 0x%x", groupOffset, dataLen)
	}
	if saclOffset > 0 && saclOffset >= dataLen {
		return nil, fmt.Errorf("invalid security descriptor: SACL offset 0x%x exceeds data length 0x%x", saclOffset, dataLen)
	}
	if daclOffset > 0 && daclOffset >= dataLen {
		return nil, fmt.Errorf("invalid security descriptor: DACL offset 0x%x exceeds data length 0x%x", daclOffset, dataLen)
	}

	// Parse Owner SID if present
	var ownerSID *sid
	if ownerOffset > 0 {
		sid, err := parseSIDBinary(data[ownerOffset:])
		if err != nil {
			return nil, fmt.Errorf("error parsing owner SID: %w", err)
		}
		ownerSID = sid
	}

	// Parse Group SID if present
	var groupSID *sid
	if groupOffset > 0 {
		sid, err := parseSIDBinary(data[groupOffset:])
		if err != nil {
			return nil, fmt.Errorf("error parsing group SID: %w", err)
		}
		groupSID = sid
	}

	// Parse DACL if present
	var dacl *acl
	if daclOffset > 0 {
		acl, err := parseACLBinary(data[daclOffset:], "D", control)
		if err != nil {
			return nil, fmt.Errorf("error parsing DACL: %w", err)
		}
		dacl = acl
	}

	// Parse SACL if present
	var sacl *acl
	if saclOffset > 0 {
		acl, err := parseACLBinary(data[saclOffset:], "S", control)
		if err != nil {
			return nil, fmt.Errorf("error parsing SACL: %w", err)
		}
		sacl = acl
	}

	return &SecurityDescriptor{
		revision:    revision,
		sbzl:        sbzl,
		control:     control,
		ownerOffset: ownerOffset,
		groupOffset: groupOffset,
		saclOffset:  saclOffset,
		daclOffset:  daclOffset,
		ownerSID:    ownerSID,
		groupSID:    groupSID,
		dacl:        dacl,
		sacl:        sacl,
	}, nil
}

// parseACEBinary takes a binary ACE and returns an ACE struct
func parseACEBinary(data []byte) (*ace, error) {
	dataLen := uint16(len(data))
	if dataLen < 16 {
		return nil, fmt.Errorf("invalid ACE: too short, got %d bytes but need at least 16 (4 for header + 4 for access mask + 8 for SID)", dataLen)
	}

	aceType := data[0]
	aceFlags := data[1]
	aceSize := binary.LittleEndian.Uint16(data[2:4])

	// Validate full ACE size fits in data provided
	if dataLen < aceSize {
		return nil, fmt.Errorf("invalid ACE: data length %d doesn't match ACE size %d", dataLen, aceSize)
	}

	accessMask := binary.LittleEndian.Uint32(data[4:8])

	sid, err := parseSIDBinary(data[8:])
	if err != nil {
		return nil, fmt.Errorf("error parsing ACE SID: %w", err)
	}

	return &ace{
		header: &aceHeader{
			aceType:  aceType,
			aceFlags: aceFlags,
			aceSize:  aceSize,
		},
		accessMask: accessMask,
		sid:        sid,
	}, nil
}

// parseACLBinary takes a binary ACL and returns an ACL struct
func parseACLBinary(data []byte, aclType string, control uint16) (*acl, error) {
	dataLength := uint16(len(data))
	if dataLength < 8 {
		return nil, fmt.Errorf("invalid ACL: too short")
	}

	aclRevision := data[0]
	sbzl := data[1]
	aclSize := binary.LittleEndian.Uint16(data[2:4])
	aceCount := binary.LittleEndian.Uint16(data[4:6])
	sbz2 := binary.LittleEndian.Uint16(data[6:8])

	var aces []ace
	offset := uint16(8)

	// Parse each ACE
	for i := uint16(0); i < aceCount; i++ {
		if offset >= aclSize {
			return nil, fmt.Errorf("invalid ACL: offset is bigger than AclSize: offset 0x%x (ACL Size: 0x%x)", offset, aclSize)
		}

		ace, err := parseACEBinary(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("error parsing ACE: %w", err)
		}

		aces = append(aces, *ace)
		offset += uint16(ace.header.aceSize)
	}

	return &acl{
		aclRevision: aclRevision,
		sbzl:        sbzl,
		aclSize:     aclSize,
		aceCount:    aceCount,
		sbz2:        sbz2,
		aclType:     aclType,
		control:     control,
		aces:        aces,
	}, nil
}

// parseSIDBinary takes a binary SID and returns a SID struct
func parseSIDBinary(data []byte) (*sid, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid SID: it must be at least 8 bytes long")
	}

	revision := data[0]
	subAuthorityCount := int(data[1])

	neededLen := 8 + (4 * subAuthorityCount)
	if len(data) < neededLen {
		return nil, fmt.Errorf("invalid SID: truncated data, got %d bytes but need %d bytes for %d sub-authorities",
			len(data), neededLen, subAuthorityCount)
	}

	if subAuthorityCount > 15 { // Maximum sub-authorities in a valid SID
		return nil, fmt.Errorf("invalid SID: too many sub-authorities (%d), maximum is 15", subAuthorityCount)
	}

	if len(data) < 8+4*subAuthorityCount {
		return nil, fmt.Errorf("invalid SID: data too short for sub-authority count")
	}

	// Parse authority (48 bits)
	authority := uint64(0)
	for i := 2; i < 8; i++ {
		authority = authority<<8 | uint64(data[i])
	}

	// Parse sub-authorities
	subAuthorities := make([]uint32, subAuthorityCount)
	for i := 0; i < subAuthorityCount; i++ {
		offset := 8 + 4*i
		subAuthorities[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	return &sid{
		revision:            revision,
		identifierAuthority: authority,
		subAuthority:        subAuthorities,
	}, nil
}
