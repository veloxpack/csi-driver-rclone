package sddl

import (
	"fmt"
	"strconv"
	"strings"
)

// wellKnownRIDs maps short names to Relative Identifiers (RIDs) for well-known security principals
// as defined in [MS-DTYP] section 2.4.2.4 Well-known SID Structures.
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/81d92bba-d22b-4a8c-908a-554ab29148ab
var wellKnownRIDs = map[string]rid{
	"LA": 500, // DOMAIN_USER_RID_ADMIN (Local Administrator)
	"LG": 501, // DOMAIN_USER_RID_GUEST (Local Guest)
}

// sidHolder represents any structure capable of containing zero or more Security Identifiers (SIDs).
//
// This interface is necessary for two main reasons:
//  1. Parsing SDDL components may result in incomplete SID parsing results.
//  2. At some point, we need to extract all complete SIDs from existing structures
//     to build the incomplete SIDs (using the domain information from complete SIDs).
//
// Implementations of this interface should provide a method to access all contained SIDs.
type sidHolder interface {
	// sids returns a slice of all SIDs contained within the implementing structure.
	sids() []sid
}

// making existing structures implement sidHolder

var _ sidHolder = &sid{}

func (s *sid) sids() []sid { // implements sidHolder
	return []sid{*s}
}

var _ sidHolder = &ace{}

func (a *ace) sids() []sid { // implements sidHolder
	return []sid{*a.sid}
}

var _ sidHolder = &acl{}

func (a *acl) sids() []sid { // implements sidHolder
	var sids []sid
	for _, ace := range a.aces {
		sids = append(sids, ace.sids()...)
	}
	return sids
}

// parseSIDStringResult represents the outcome of a SID parsing operation.
//
// This interface can represent either:
//   - A complete SID structure
//   - An incomplete SID for domain-specific Relative Identifiers (RIDs)
//     where domain information is missing (e.g., S-1-5-21-<domain>-<rid>)
//
// Implementations must provide a method to convert the result into a full SID,
// potentially using contextual information from previously parsed SIDs.
type parseSIDStringResult interface {
	sidHolder // parseSIDStringResult implements sidHolder, incomplete results return empty slice

	// toSID converts the result into a full SID.
	//
	// It uses contextual information from previously parsed SIDs if necessary.
	// For incomplete SIDs (e.g., RIDs without domain information), it attempts to
	// extract domain information from previousSIDs. If previousSIDs is empty and
	// the SID is incomplete, this method will return an error.
	//
	// Parameters:
	//   - previousSIDs: A slice of previously parsed SIDs to provide context
	//
	// Returns:
	//   - *sid: A pointer to the complete SID structure
	//   - error: An error if the conversion fails
	toSID(previousSIDs []sid) (*sid, error)
}

func (s *sid) toSID(previousSIDs []sid) (*sid, error) {
	// sid structure is a valid parseSIDStringResult and represents a complete SID
	return s, nil
}

// rid represents a Relative Identifier (RID), which is the last sub-authority of a Security Identifier (SID).
// It is incomplete on its own and requires domain information from a complete SID to form a full SID.
// RIDs are typically used in domain environments to uniquely identify users, groups, or other security principals.
type rid uint32

func (r rid) toSID(previousSIDs []sid) (*sid, error) {
	if len(previousSIDs) == 0 {
		return nil, ErrMissingDomainInformation
	}

	s, err := r.complete(previousSIDs[0])
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r rid) sids() []sid {
	return []sid{}
}

// complete converts a Relative Identifier (RID) into a complete SID by combining it with the information from an existing SID.
// It uses the domain information from the provided SID and appends the RID as the last sub-authority.
//
// Parameters:
//   - s: An existing SID to provide the domain information
//
// Returns:
//   - *sid: A pointer to a new, complete SID that includes the RID
//   - error: If the sid does not contain sub authorities (first sub-authority is required)
func (r rid) complete(s sid) (*sid, error) {
	if len(s.subAuthority) == 0 {
		return nil, ErrMissingSubAuthorities
	}

	firstSubAuthority := s.subAuthority[0]
	domain := s.Domain()

	var subAuthorities []uint32
	subAuthorities = append(subAuthorities, firstSubAuthority)
	subAuthorities = append(subAuthorities, domain...)
	subAuthorities = append(subAuthorities, uint32(r))

	return &sid{
		revision:            s.revision,
		identifierAuthority: s.identifierAuthority,
		subAuthority:        subAuthorities,
	}, nil
}

// parseACEStringResult represents the outcome of an ACE parsing operation.
// It mimics the ACE structure (ace) but instead of a sid, it contains a parseSIDStringResult.
type parseACEStringResult struct {
	// header contains the ACE header information
	header *aceHeader
	// accessMask specifies the access rights controlled by the ACE
	accessMask uint32
	// sid represents the Security Identifier (SID) associated with this ACE
	sid parseSIDStringResult
}

func (a *parseACEStringResult) sids() []sid {
	return a.sid.sids()
}

// toACE converts a parseACEStringResult to a complete ACE structure.
// It resolves any incomplete SID information using the provided previousSIDs.
//
// Parameters:
//   - previousSIDs: A slice of previously parsed SIDs to provide context for incomplete SIDs
//
// Returns:
//   - *ace: A pointer to the complete ACE structure
//   - error: An error if the conversion fails, particularly if SID resolution fails
func (a *parseACEStringResult) toACE(previousSIDs []sid) (*ace, error) {
	sid, err := a.sid.toSID(previousSIDs)
	if err != nil {
		return nil, err
	}

	// Calculate the total size of the ACE
	// Size = sizeof(ACE_HEADER) + sizeof(ACCESS_MASK) + size of the SID
	// SID size = 8 + (4 * number of sub-authorities)
	sidSize := 8 + (4 * len(sid.subAuthority))
	aceSize := 4 + 4 + sidSize // 4 (header) + 4 (access mask) + sidSize
	a.header.aceSize = uint16(aceSize)

	return &ace{
		header:     a.header,
		accessMask: a.accessMask,
		sid:        sid,
	}, nil
}

// parseACLStringResult represents the outcome of an ACL parsing operation.
// It mimics the ACL structure (acl) but instead of a slice of aces, it contains a slice of parseACEStringResult.
type parseACLStringResult struct {
	// aclRevision is the revision level of the ACL structure
	aclRevision byte
	// sbzl is a reserved field (should be zero)
	sbzl byte
	// aclSize is the size, in bytes, of the ACL structure
	aclSize uint16
	// aceCount is the number of ACEs in the ACL
	aceCount uint16
	// sbz2 is a reserved field (should be zero)
	sbz2 uint16
	// aclType indicates whether this is a DACL or SACL
	aclType string
	// control contains ACL control flags
	control uint16
	// aces is a slice of parsed ACE results
	aces []parseACEStringResult
}

func (a *parseACLStringResult) sids() []sid {
	var sids []sid
	for _, ace := range a.aces {
		sids = append(sids, ace.sids()...)
	}
	return sids
}

// toACL converts a parseACLStringResult to a complete ACL structure.
// It resolves any incomplete SID information in the ACEs using the provided previousSIDs.
//
// Parameters:
//   - previousSIDs: A slice of previously parsed SIDs to provide context for incomplete SIDs in ACEs
//
// Returns:
//   - *acl: A pointer to the complete ACL structure
//   - error: An error if the conversion fails, particularly if SID resolution fails in any ACE
func (a *parseACLStringResult) toACL(previousSIDs []sid) (*acl, error) {
	var aces []ace
	for _, ace := range a.aces {
		ace, err := ace.toACE(previousSIDs)
		if err != nil {
			return nil, err
		}
		aces = append(aces, *ace)
	}

	// Calculate total ACL size
	totalSize := 8 // ACL header size
	for _, ace := range aces {
		totalSize += int(ace.header.aceSize)
	}
	a.aclSize = uint16(totalSize)

	return &acl{
		aclRevision: a.aclRevision,
		sbzl:        a.sbzl,
		aclSize:     a.aclSize,
		aceCount:    a.aceCount,
		sbz2:        a.sbz2,
		aclType:     a.aclType,
		control:     a.control,
		aces:        aces,
	}, nil
}

// FromString parses a security descriptor string in SDDL format.
// The format is: "O:owner_sidG:group_sidD:dacl_flagsS:sacl_flags"
// where each component is optional.
//
// Examples:
// - "O:SYG:BAD:(A;;FA;;;SY)"            - Owner: SYSTEM, Group: BUILTIN\Administrators, DACL with full access for SYSTEM
// - "O:SYG:SYD:PAI(A;;FA;;;SY)"         - Protected auto-inherited DACL
// - "O:SYG:SYD:(A;;FA;;;SY)S:(AU;SA;FA;;;SY)" - With both DACL and SACL
func FromString(s string) (*SecurityDescriptor, error) {
	// Initialize security descriptor with self-relative flag
	sd := &SecurityDescriptor{
		revision: 1,
		control:  seSelfRelative | seOwnerDefaulted | seGroupDefaulted | seDACLDefaulted | seSACLDefaulted, // All components are defaulted unless they are present
	}

	// Empty string is valid - returns a security descriptor with defaults set
	if s == "" {
		return sd, nil
	}

	remaining := s
	var err error

	// parsing results
	var (
		completeSIDs []sid
		ownerSID     parseSIDStringResult
		groupSID     parseSIDStringResult
		dacl         *parseACLStringResult
		sacl         *parseACLStringResult
	)

	// Parse each component in order if present
	// The order doesn't technically matter, so, we are going to keep a list of pending components to parse
	// and remove them as we go
	pendingComponents := []string{"O:", "G:", "D:", "S:"}
	removePendingComponent := func(component string) {
		for i, c := range pendingComponents {
			if c == component {
				pendingComponents = append(pendingComponents[:i], pendingComponents[i+1:]...)
				break
			}
		}
	}

	// If there is data, then, at least one component must be present
	if findNextComponent(remaining, pendingComponents...) == -1 {
		return nil, fmt.Errorf("no components found in security descriptor")
	}

	// Parse each component regardless of their order, as long as there are remaining characters and pending components
	for len(pendingComponents) > 0 && len(remaining) > 0 {
		switch {
		case strings.HasPrefix(remaining, "O:"):
			// remove O: prefix
			remaining = remaining[2:]
			removePendingComponent("O:")
			ownerSID, remaining, err = parseSIDComponent(remaining, pendingComponents...)
			if err != nil {
				return nil, fmt.Errorf("error parsing owner SID: %w", err)
			}
			sd.control ^= seOwnerDefaulted

		case strings.HasPrefix(remaining, "G:"):
			// remove G: prefix
			remaining = remaining[2:]
			removePendingComponent("G:")
			groupSID, remaining, err = parseSIDComponent(remaining, pendingComponents...)
			if err != nil {
				return nil, fmt.Errorf("error parsing group SID: %w", err)
			}
			sd.control ^= seGroupDefaulted

		case strings.HasPrefix(remaining, "D:"):
			// remove D: prefix
			remaining = remaining[2:]
			removePendingComponent("D:")
			dacl, remaining, err = parseACLComponent("D", remaining, pendingComponents...)
			if err != nil {
				return nil, fmt.Errorf("error parsing DACL: %w", err)
			}
			sd.control ^= seDACLDefaulted
			sd.control |= seDACLPresent

		case strings.HasPrefix(remaining, "S:"):
			// remove S: prefix
			remaining = remaining[2:]
			removePendingComponent("S:")
			sacl, remaining, err = parseACLComponent("S", remaining, pendingComponents...)
			if err != nil {
				return nil, fmt.Errorf("error parsing SACL: %w", err)
			}
			sd.control ^= seSACLDefaulted
			sd.control |= seSACLPresent
		}
	}

	// If there's anything left unparsed, it's an error
	if remaining != "" {
		return nil, fmt.Errorf("unexpected content after parsing: %s", remaining)
	}

	// convert parsed result components into final structures
	if ownerSID != nil {
		completeSIDs = append(completeSIDs, ownerSID.sids()...)
	}
	if groupSID != nil {
		completeSIDs = append(completeSIDs, groupSID.sids()...)
	}
	if dacl != nil {
		completeSIDs = append(completeSIDs, dacl.sids()...)
	}
	if sacl != nil {
		completeSIDs = append(completeSIDs, sacl.sids()...)
	}

	// Remove generic (well-known) SIDs from completeSIDs because they do not give the appropriate domain
	for i := len(completeSIDs) - 1; i >= 0; i-- {
		if completeSIDs[i].isGeneric() {
			completeSIDs = append(completeSIDs[:i], completeSIDs[i+1:]...)
		}
	}

	// Resolve incomplete SIDs in the DACL and SACL
	if dacl != nil {
		sd.dacl, err = dacl.toACL(completeSIDs)
		if err != nil {
			return nil, err
		}
	}
	if sacl != nil {
		sd.sacl, err = sacl.toACL(completeSIDs)
		if err != nil {
			return nil, err
		}
	}
	if ownerSID != nil {
		sd.ownerSID, err = ownerSID.toSID(completeSIDs)
		if err != nil {
			return nil, err
		}
	}
	if groupSID != nil {
		sd.groupSID, err = groupSID.toSID(completeSIDs)
		if err != nil {
			return nil, err
		}
	}

	// update control flags based on ACLs
	if sd.dacl != nil {
		// Update control flags based on DACL flags
		if sd.dacl.control&seDACLProtected != 0 {
			sd.control |= seDACLProtected
		}
		if sd.dacl.control&seDACLAutoInherited != 0 {
			sd.control |= seDACLAutoInherited
		}
		if sd.dacl.control&seDACLAutoInheritRe != 0 {
			sd.control |= seDACLAutoInheritRe
		}
	}
	if sd.sacl != nil {
		// Update control flags based on SACL flags
		if sd.sacl.control&seSACLProtected != 0 {
			sd.control |= seSACLProtected
		}
		if sd.sacl.control&seSACLAutoInherited != 0 {
			sd.control |= seSACLAutoInherited
		}
		if sd.sacl.control&seSACLAutoInheritRe != 0 {
			sd.control |= seSACLAutoInheritRe
		}
	}

	// Adjust ACL's control flags once they are fully computed
	if sd.dacl != nil {
		sd.dacl.control = sd.control
	}
	if sd.sacl != nil {
		sd.sacl.control = sd.control
	}

	return sd, nil
}

func parseSIDComponent(s string, nextMarkers ...string) (sid parseSIDStringResult, remaining string, err error) {
	// Find the next component marker (G:, D:, or S:)
	sidEnd := findNextComponent(s, nextMarkers...)
	if sidEnd == -1 {
		sidEnd = len(s)
	}

	// Parse the SID string
	sid, err = parseSIDString(s[:sidEnd])
	if err != nil {
		return nil, "", fmt.Errorf("invalid SID: %w", err)
	}

	return sid, s[sidEnd:], nil
}

func parseACLComponent(aclType, s string, nextMarkers ...string) (aclr *parseACLStringResult, remaining string, err error) {
	// Find the next marker (if any)
	aclEnd := len(s)
	if len(nextMarkers) > 0 {
		nextMarkerIndex := findNextComponent(s, nextMarkers...)
		if nextMarkerIndex != -1 {
			aclEnd = nextMarkerIndex
		}
	}

	// Parse the ACL string
	aclr, err = parseACLString(aclType, s[:aclEnd])
	if err != nil {
		return nil, "", fmt.Errorf("invalid ACL: %w", err)
	}

	return aclr, s[aclEnd:], nil
}

// findNextComponent looks for the next component marker given in arguments
// Returns the index of the next component or -1 if none found
func findNextComponent(s string, markers ...string) int {
	minIndex := -1
	for _, marker := range markers {
		if idx := strings.Index(s, marker); idx != -1 {
			if minIndex == -1 || idx < minIndex {
				minIndex = idx
			}
		}
	}

	return minIndex
}

// parseAccessMask converts an access mask string to its corresponding uint32 value
func parseAccessMask(maskStr string) (uint32, error) {
	// Check well-known access masks first
	if value, ok := reverseWellKnownAccessMasks[maskStr]; ok {
		return value, nil
	}

	// If not a well-known mask, try to parse as hexadecimal
	if strings.HasPrefix(maskStr, "0x") {
		value, err := strconv.ParseUint(maskStr[2:], 16, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid hexadecimal access mask: %s", maskStr)
		}
		return uint32(value), nil
	}

	// If not a hexadecimal, try to use two-letter codes

	var components []string
	var idx int
	for idx < len(maskStr) {
		components = append(components, maskStr[idx:idx+2])
		idx += 2
	}

	mask, remaining := composeAccessMask(components)
	if len(remaining) == 0 {
		return mask, nil
	}

	return 0, fmt.Errorf("unknown access mask: %s", maskStr)
}

// parseACEString parses an ACE string in the format "(type;flags;rights;;;sid)" into an ACE structure
// Example: "(A;;FA;;;SY)" which represents:
// - Type: A (ACCESS_ALLOWED_ACE_TYPE)
// - Flags: (none)
// - Rights: FA (Full Access)
// - SID: SY (Local System)
func parseACEString(aceStr string) (*parseACEStringResult, error) {
	// Validate basic string format
	if len(aceStr) < 2 || !strings.HasPrefix(aceStr, "(") || !strings.HasSuffix(aceStr, ")") {
		return nil, fmt.Errorf("invalid ACE string format: must be enclosed in parentheses")
	}

	// Remove parentheses and split into components
	parts := strings.Split(aceStr[1:len(aceStr)-1], ";")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid ACE string format: expected 6 components separated by semicolons")
	}

	// Parse ACE type
	aceType, err := parseACEType(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid ACE type: %w", err)
	}

	// Parse ACE flags with type validation
	aceFlags, err := parseFlagsForACEType(parts[1], aceType)
	if err != nil {
		return nil, fmt.Errorf("invalid ACE flags: %w", err)
	}

	// Parse access mask
	accessMask, err := parseAccessMask(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid access mask: %w", err)
	}

	// Parse SID (parts[3] and parts[4] are object type and inherited object type, which we ignore)
	sid, err := parseSIDString(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid SID: %w", err)
	}

	ace := &parseACEStringResult{
		header: &aceHeader{
			aceType:  aceType,
			aceFlags: aceFlags,
		},
		accessMask: accessMask,
		sid:        sid,
	}

	return ace, nil
}

// parseACEType converts an ACE type string to its corresponding byte value
// The valid types are:
// - A (ACCESS_ALLOWED_ACE_TYPE): allows access to the object
// - D (ACCESS_DENIED_ACE_TYPE): denies access to the object
// - AU (SYSTEM_AUDIT_ACE_TYPE): specifies a system audit ACE
// - AL (SYSTEM_ALARM_ACE_TYPE): specifies a system alarm ACE
// - OA (ACCESS_ALLOWED_OBJECT_ACE_TYPE): specifies an object-specific access ACE
func parseACEType(typeStr string) (byte, error) {
	// First check well-known string representations
	switch typeStr {
	case "A":
		return accessAllowedACEType, nil
	case "D":
		return accessDeniedACEType, nil
	case "AU":
		return systemAuditACEType, nil
	case "AL":
		return systemAlarmACEType, nil
	case "OA":
		return accessAllowedObjectACEType, nil
	}

	// If not a well-known type, try to parse as hexadecimal
	// The format should be "0xNN" where NN is a hex number
	if strings.HasPrefix(typeStr, "0x") {
		value, err := strconv.ParseUint(typeStr[2:], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("invalid hexadecimal ACE type: %s", typeStr)
		}
		return byte(value), nil
	}

	return 0, fmt.Errorf("invalid ACE type: %s (must be a known type or hexadecimal value)", typeStr)
}

// parseACLFlags splits a flag string into individualn ACL flags
// Example: "PAI" becomes []string{"P", "AI"}
//
// The ACL Control Flags in SDDL String Format are:
//
// Single-letter flags:
//
//	P - Protected
//	    Prevents the ACL from being modified by inheritable ACEs.
//	    The ACL is protected from inheritance flowing down from parent containers.
//	R - Read-Only
//	    Marks the ACL as read-only, preventing any modifications.
//	    This is often used for system-managed ACLs.
//
// Two-letter flags:
//
//	AI - Auto-Inherited
//	    Indicates the ACL was created through inheritance.
//	    Appears when the ACL contains entries inherited from a parent object.
//	AR - Auto-Inherit Required
//	    Forces child objects to inherit this ACL.
//	    When set, ensures all child objects must process inherited permissions.
//	NO - No Inheritance
//	    Explicitly excludes inheritable ACEs from being considered.
//	    Blocks inheritance without changing the inherited ACEs themselves.
//	IO - Inherit Only
//	    Specifies the ACL should only be used for inheritance purposes.
//	    The ACL is not used for access checks on the current object.
//
// These flags can be combined in any order after the ACL type identifier:
// - For DACLs: "D:[flags]", e.g., "D:PAI", "D:AINO"
// - For SACLs: "S:[flags]", e.g., "S:PAR", "S:ARNO"
//
// The ordering of combined flags does not affect their meaning:
// "D:AINO" is equivalent to "D:NOAI"
func parseACLFlags(s string) ([]string, error) {
	var flags []string
	for i := 0; i < len(s); {
		code1 := s[i : i+1]
		code2 := ""
		if i+1 < len(s) {
			code2 = s[i : i+2]
		}

		// Check for two-character flags first
		switch code2 {
		case "AI", "AR", "NO", "IO":
			flags = append(flags, code2)
			i += 2
		default:
			// Check for single-character flags
			switch code1 {
			case "P", "R":
				flags = append(flags, code1)
				i++
			default:
				return nil, fmt.Errorf("invalid flag: %q", s[i])
			}
		}
	}
	return flags, nil
}

// parseACLString parses an ACL string representation into an ACL structure.
// The ACL string format follows the Security Descriptor String Format (SDDL).
// Parameters:
//   - aclType: Either "D" for DACL or "S" for SACL
//   - s: The ACL string to parse, which may include:
//   - Optional flags (e.g., "PAI" for Protected and AutoInherited)
//   - One or more ACEs enclosed in parentheses
//
// Examples:
//   - "D:(A;;FA;;;SY)"           // DACL with a single ACE
//   - "S:PAI(AU;SA;FA;;;SY)"     // Protected auto-inherited SACL with an audit ACE
//   - "D:(A;;FA;;;SY)(D;;FR;;;WD)" // DACL with two ACEs
func parseACLString(aclType, s string) (*parseACLStringResult, error) {
	// Determine ACL type from prefix
	var baseControl uint16
	switch aclType {
	case "D":
		baseControl = seDACLPresent
	case "S":
		baseControl = seSACLPresent
	default:
		return nil, fmt.Errorf("invalid ACL type: must be either 'D' or 'S'")
	}

	// Parse flags if present (before the first ACE)
	var control uint16 = baseControl
	var flags []string
	aceStart := 0

	// Look for flags section (between : and first parenthesis)
	if len(s) > 0 && s[0] != '(' {
		flagEnd := strings.Index(s, "(")
		if flagEnd == -1 {
			if strings.Contains(s, ")") {
				return nil, fmt.Errorf("invalid ACL format: missing opening parenthesis")
			}
			flagEnd = len(s)
		}
		ff, err := parseACLFlags(s[:flagEnd])
		if err != nil {
			return nil, fmt.Errorf("error parsing flags: %w", err)
		}
		flags = ff
		aceStart = flagEnd
	}

	// Update control flags based on parsed flags
	// Note: other flags such as NO, IO, etc. are ignored because they do not have a corresponding control flag
	for _, flag := range flags {
		switch flag {
		case "P":
			if aclType == "D" {
				control |= seDACLProtected
			} else {
				control |= seSACLProtected
			}
		case "AI":
			if aclType == "D" {
				control |= seDACLAutoInherited
			} else {
				control |= seSACLAutoInherited
			}
		case "AR":
			if aclType == "D" {
				control |= seDACLAutoInheritRe
			} else {
				control |= seSACLAutoInheritRe
			}
		case "R":
			if aclType == "D" {
				control |= seDACLDefaulted
			} else {
				control |= seSACLDefaulted
			}
		}
	}

	// Parse ACEs
	var aces []parseACEStringResult
	remaining := s[aceStart:]

	// Handle empty ACL (no ACEs)
	if len(remaining) == 0 {
		return &parseACLStringResult{
			aclRevision: 2,
			aclSize:     8, // Size of empty ACL (just header)
			aclType:     aclType,
			control:     control,
		}, nil
	}

	// Extract each ACE string (enclosed in parentheses)
	for len(remaining) > 0 {
		if remaining[0] != '(' {
			return nil, fmt.Errorf("invalid ACE format: expected '(' but got %q", remaining[0])
		}

		// Find closing parenthesis
		closePos := strings.Index(remaining, ")")
		if closePos == -1 {
			return nil, fmt.Errorf("invalid ACE format: missing closing parenthesis")
		}

		// Parse individual ACE
		aceStr := remaining[:closePos+1]
		ace, err := parseACEString(aceStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing ACE %q: %w", aceStr, err)
		}

		aces = append(aces, *ace)
		remaining = remaining[closePos+1:]
	}

	// Create and return the ACL structure
	return &parseACLStringResult{
		aclRevision: 2,
		sbzl:        0,
		aceCount:    uint16(len(aces)),
		sbz2:        0,
		aclType:     aclType,
		control:     control,
		aces:        aces,
	}, nil
}

// parseFlagsForACEType converts an ACE flags string to its corresponding byte value,
// validating that the flags are appropriate for the given ACE type
func parseFlagsForACEType(flagsStr string, aceType byte) (byte, error) {
	if flagsStr == "" {
		return 0, nil
	}

	var flags byte
	var hasAuditFlags bool

	// Process flags in pairs (each flag is 2 characters)
	for i := 0; i < len(flagsStr); i += 2 {
		if i+2 > len(flagsStr) {
			return 0, fmt.Errorf("invalid flag format at position %d", i)
		}

		flag := flagsStr[i : i+2]
		switch flag {
		// Inheritance flags - valid for all ACE types
		case "CI":
			flags |= containerInheritACE
		case "OI":
			flags |= objectInheritACE
		case "NP":
			flags |= noPropagateInheritACE
		case "IO":
			flags |= inheritOnlyACE
		case "ID":
			flags |= inheritedACE
		// Audit flags - only valid for SYSTEM_AUDIT_ACE_TYPE
		case "SA", "FA":
			hasAuditFlags = true
			if aceType != systemAuditACEType {
				return 0, fmt.Errorf("audit flags (SA/FA) are only valid for audit ACEs")
			}
			if flag == "SA" {
				flags |= successfulAccessACE
			} else {
				flags |= failedAccessACE
			}
		default:
			return 0, fmt.Errorf("unknown flag: %s", flag)
		}
	}

	// Validate that audit ACEs have at least one audit flag
	if aceType == systemAuditACEType && !hasAuditFlags {
		return 0, fmt.Errorf("audit ACEs must specify at least one audit flag (SA/FA)")
	}

	return flags, nil
}

// parseSIDString parses a string SID representation into a SID structure
func parseSIDString(s string) (parseSIDStringResult, error) {
	// First, check if it's a well-known RID abbreviation
	// hence this parsing will result in an incomplete SID
	if r, ok := wellKnownRIDs[s]; ok {
		return r, nil
	}

	// Second, check if it's a well-known SID abbreviation
	if fullSid, ok := reverseWellKnownSids[s]; ok {
		s = fullSid
	}

	// If it doesn't start with "S-", it's invalid
	if !strings.HasPrefix(s, "S-") {
		return nil, fmt.Errorf("%w: must start with S-", ErrInvalidSIDFormat)
	}

	// Split the SID string into components
	parts := strings.Split(s[2:], "-") // Skip "S-" prefix
	if len(parts) < 2 {
		return nil, fmt.Errorf("%w: insufficient components", ErrInvalidSIDFormat)
	}

	// Parse revision
	revision, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRevision, err)
	}
	if revision != 1 {
		return nil, fmt.Errorf("%w: got %d, want 1", ErrInvalidRevision, revision)
	}

	// Parse authority - can be decimal or hex (with 0x prefix)
	var authority uint64
	authStr := parts[1]
	if strings.HasPrefix(strings.ToLower(authStr), "0x") {
		// Parse hexadecimal authority
		authority, err = strconv.ParseUint(authStr[2:], 16, 48)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid hex value %v", ErrInvalidAuthority, err)
		}
	} else {
		// Parse decimal authority
		authority, err = strconv.ParseUint(authStr, 10, 48)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid decimal value %v", ErrInvalidAuthority, err)
		}
	}

	// Additional validation for authority value
	if authority >= 1<<48 {
		return nil, fmt.Errorf("%w: value %d exceeds maximum of 2^48-1", ErrInvalidAuthority, authority)
	}

	// Parse sub-authorities
	subAuthCount := len(parts) - 2 // Subtract revision and authority parts
	if subAuthCount > 15 {
		return nil, fmt.Errorf("%w: got %d, maximum is 15", ErrTooManySubAuthorities, subAuthCount)
	}

	subAuthorities := make([]uint32, subAuthCount)
	for i := 0; i < subAuthCount; i++ {
		sa, err := strconv.ParseUint(parts[i+2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid sub-authority at position %d: %v",
				ErrInvalidSubAuthority, i, err)
		}
		subAuthorities[i] = uint32(sa)
	}

	return &sid{
		revision:            byte(revision),
		identifierAuthority: authority,
		subAuthority:        subAuthorities,
	}, nil
}
