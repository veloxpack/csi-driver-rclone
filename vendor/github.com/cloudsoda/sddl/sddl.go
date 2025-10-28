package sddl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Define common errors
var (
	ErrInvalidAuthority         = errors.New("invalid authority value")
	ErrInvalidRevision          = errors.New("invalid SID revision")
	ErrInvalidSIDFormat         = errors.New("invalid SID format")
	ErrInvalidSubAuthority      = errors.New("invalid sub-authority value")
	ErrMissingDomainInformation = errors.New("missing domain information")
	ErrMissingSubAuthorities    = errors.New("missing sub-authorities")
	ErrTooManySubAuthorities    = errors.New("too many sub-authorities")
)

// constants for SECURITY_DESCRIPTOR parsing
//
// Defaulted refers to the situation where a security descriptor is taken from somewhere else,
// usually from the parent object. This is a common situation where a file inherits its permissions
// from the parent directory. In this case, the file's DACL is defaulted to the DACL of the directory.
//
// Inherited refers to the situation where a security descriptor is taken from somewhere else,
// usually from the parent object, but it is not the same as the parent object's security descriptor.
// It is a copy of the parent object's security descriptor, but it is applied to the child object.
//
// Protected refers to the situation where the security descriptor is protected against
// inheritance. This means that the security descriptor is not inherited by any child
// objects, and any changes to the security descriptor will not affect the child objects.
//
// Auto-inherited refers to the situation where a security descriptor is automatically
// inherited from the parent object. This means that the security descriptor is copied
// from the parent object to the child object when the child object is created.
//
// Auto-inherited required (RE) refers to the situation where a security descriptor is
// automatically inherited from the parent object, and the child object must inherit the
// security descriptor. This means that the child object cannot override the security
// descriptor of the parent object.
//
// # See
//
//   - https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-control
//   - https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-ace_header
//   - https://docs.microsoft.com/en-us/windows/win32/secauthz/access-mask-format
const (
	// Control flags

	// seOwnerDefaulted - Owner is defaulted to current owner (SE_OWNER_DEFAULTED)
	seOwnerDefaulted = 0x0001
	// seGroupDefaulted - Group is defaulted to current group (SE_GROUP_DEFAULTED)
	seGroupDefaulted = 0x0002
	// seDACLPresent - Indicates that a DACL is present (SE_DACL_PRESENT)
	seDACLPresent = 0x0004
	// seDACLDefaulted - Indicates that DACL is defaulted (SE_DACL_DEFAULTED)
	seDACLDefaulted = 0x0008
	// seSACLPresent - SACL is present (SE_SACL_PRESENT)
	seSACLPresent = 0x0010
	// seSACLDefaulted - SACL is defaulted (SE_SACL_DEFAULTED)
	seSACLDefaulted = 0x0020
	// seDACLTrusted - DACL is trusted (SE_DACL_TRUSTED)
	// In this context, 'trusted' means that the DACL was set explicitly by a user or an application,
	// and should not be modified by the system.
	seDACLTrusted = 0x0040
	// seServerSecurity - Server security (SE_SERVER_SECURITY)
	// This flag is set when the security descriptor is from a server object, such as a file share.
	seServerSecurity = 0x0080
	// seDACLAutoInheritRe - Auto-inherit parent DACL (SE_DACL_AUTO_INHERIT_RE)
	seDACLAutoInheritRe = 0x0100
	// seSACLAutoInheritRe - Auto-inherit parent SACL (SE_SACL_AUTO_INHERIT_RE)
	seSACLAutoInheritRe = 0x0200
	// seDACLAutoInherited - Auto-inherited DACL (SE_DACL_AUTO_INHERITED)
	seDACLAutoInherited = 0x0400
	// seSACLAutoInherited - Auto-inherited SACL (SE_SACL_AUTO_INHERITED)
	seSACLAutoInherited = 0x0800
	// seDACLProtected - DACL is protected (SE_DACL_PROTECTED)
	seDACLProtected = 0x1000
	// seSACLProtected - SACL is protected (SE_SACL_PROTECTED)
	seSACLProtected = 0x2000
	// seResourceManagerControlValid - Resource manager control is valid (SE_RESOURCE_MANAGER_CONTROL_VALID)
	// This flag is set when the resource manager has verified that the security descriptor is valid.
	// It is used by the system to ensure that the security descriptor was set by a trusted entity.
	seResourceManagerControlValid = 0x4000
	// seSelfRelative - Self relative flag which means the information is packed in a contiguous region of memory (SE_SELF_RELATIVE)
	seSelfRelative = 0x8000

	// ACE types

	// accessAllowedACEType - Access allowed (ACCESS_ALLOWED_ACE_TYPE)
	accessAllowedACEType = 0x0
	// accessDeniedACEType - Access denied (ACCESS_DENIED_ACE_TYPE)
	accessDeniedACEType = 0x1
	// systemAuditACEType - System audit (SYSTEM_AUDIT_ACE_TYPE)
	// This ACE type is used to specify system-level auditing for an object.
	// It allows the system to track all access to the object and generate an audit log entry.
	systemAuditACEType = 0x2
	// systemAlarmACEType - System alarm (SYSTEM_ALARM_ACE_TYPE)
	// This ACE type is used to specify system-level alarms for an object.
	// It allows the system to generate alarms in response to access to the object.
	systemAlarmACEType = 0x3
	// accessAllowedObjectACEType - Access allowed object (ACCESS_ALLOWED_OBJECT_ACE_TYPE)
	accessAllowedObjectACEType = 0x5

	// ACE flags

	// objectInheritACE - Object inherit (OBJECT_INHERIT_ACE)
	// This flag is set when the ACE is inherited by objects of the same type as the object being modified.
	objectInheritACE = 0x01
	// containerInheritACE - Container inherit (CONTAINER_INHERIT_ACE)
	// This flag is set when the ACE is inherited by objects of a different type than the object being modified.
	containerInheritACE = 0x02
	// noPropagateInheritACE - No propagate inherit (NO_PROPAGATE_INHERIT_ACE)
	// This flag is set when the ACE is inherited by objects of a different type than the object being modified.
	noPropagateInheritACE = 0x04
	// inheritOnlyACE - Inherit only (INHERIT_ONLY_ACE)
	// This flag is set when the ACE is inherited by objects of a different type than the object being modified.
	inheritOnlyACE = 0x08
	// inheritedACE - Inherited (INHERITED_ACE)
	inheritedACE = 0x10
	// successfulAccessACE - Successful access (SUCCESSFUL_ACCESS_ACE)
	// This flag is set when the ACE type is ACCESS_ALLOWED_ACE_TYPE and the access is successful.
	successfulAccessACE = 0x40
	// failedAccessACE - Failed access (FAILED_ACCESS_ACE)
	// This flag is set when the ACE type is ACCESS_DENIED_ACE_TYPE and the access is denied.
	failedAccessACE = 0x80
)

// wellKnownSids maps short SID names to their full string representation as
// documented in the Microsoft documentation: https://docs.microsoft.com/en-us/windows/win32/secauthz/well-known-sids
var wellKnownSids = map[string]string{
	"S-1-0-0":      "NULL",
	"S-1-1-0":      "WD", // Everyone
	"S-1-2-0":      "LG", // Local GROUP
	"S-1-3-0":      "CC", // CREATOR CREATOR
	"S-1-3-1":      "CO", // CREATOR OWNER
	"S-1-3-2":      "CG", // CREATOR GROUP
	"S-1-3-3":      "OW", // OWNER RIGHTS
	"S-1-5-1":      "DU", // DIALUP
	"S-1-5-2":      "AN", // NETWORK
	"S-1-5-3":      "BT", // BATCH
	"S-1-5-4":      "IU", // INTERACTIVE
	"S-1-5-6":      "SU", // SERVICE
	"S-1-5-7":      "AS", // ANONYMOUS
	"S-1-5-8":      "PS", // PROXY
	"S-1-5-9":      "ED", // ENTERPRISE DOMAIN CONTROLLERS
	"S-1-5-10":     "SS", // SELF
	"S-1-5-11":     "AU", // Authenticated Users
	"S-1-5-12":     "RC", // RESTRICTED CODE
	"S-1-5-18":     "SY", // LOCAL SYSTEM
	"S-1-5-32-544": "BA", // BUILTIN\Administrators
	"S-1-5-32-545": "BU", // BUILTIN\Users
	"S-1-5-32-546": "BG", // BUILTIN\Guests
	"S-1-5-32-547": "PU", // BUILTIN\Power Users
	"S-1-5-32-548": "AO", // BUILTIN\Account Operators
	"S-1-5-32-549": "SO", // BUILTIN\Server Operators
	"S-1-5-32-550": "PO", // BUILTIN\Print Operators
	"S-1-5-32-551": "BO", // BUILTIN\Backup Operators
	"S-1-5-32-552": "RE", // BUILTIN\Replicator
	"S-1-5-32-554": "RU", // BUILTIN\Pre-Windows 2000 Compatible Access
	"S-1-5-32-555": "RD", // BUILTIN\Remote Desktop Users
	"S-1-5-32-556": "NO", // BUILTIN\Network Configuration Operators
	"S-1-5-64-10":  "AA", // Administrator Access
	"S-1-5-64-14":  "RA", // Remote Access
	"S-1-5-64-21":  "OA", // Operation Access
}

// accessMaskComponents maps permission codes to their bit values
var accessMaskComponents = map[string]uint32{
	// Generic Rights (0xF0000000)
	"GA": 0x10000000, // Generic All
	"GX": 0x20000000, // Generic Execute
	"GW": 0x40000000, // Generic Write
	"GR": 0x80000000, // Generic Read

	// ??
	"MA": 0x02000000, // Maximum Allowed
	"AS": 0x01000000, // Access System Security

	// Standard Rights (0x001F0000)
	"SY": 0x00100000, // Synchronize
	"WO": 0x00080000, // Write Owner
	"WD": 0x00040000, // Write DAC
	"RC": 0x00020000, // Read Control
	"SD": 0x00010000, // Delete

	// Directory Service Object Access Rights (0x0000FFFF)
	"CR": 0x00000100, // Control Access
	"LO": 0x00000080, // List Object
	"DT": 0x00000040, // Delete Tree
	"WP": 0x00000020, // Write Property
	"RP": 0x00000010, // Read Property
	"SW": 0x00000008, // Self Write
	"LC": 0x00000004, // List Children
	"DC": 0x00000002, // Delete Child
	"CC": 0x00000001, // Create Child
}

// WellKnownAccessMasks maps common combined access masks to their string representations
var wellKnownAccessMasks = map[uint32]string{
	0x001f01ff: "FA", // File All (STANDARD_RIGHTS_REQUIRED | SYNCHRONIZE | 0x1FF)
	0x00120089: "FR", // File Read (READ_CONTROL | FILE_READ_DATA | FILE_READ_ATTRIBUTES | FILE_READ_EA | SYNCHRONIZE)
	0x00120116: "FW", // File Write (READ_CONTROL | FILE_WRITE_DATA | FILE_WRITE_ATTRIBUTES | FILE_WRITE_EA | FILE_APPEND_DATA | SYNCHRONIZE)
	0x001200a0: "FX", // File Execute (READ_CONTROL | FILE_READ_ATTRIBUTES | FILE_EXECUTE | SYNCHRONIZE)
}

// reversedAccessMaskComponents maps access mask values to their short names
var reversedAccessMaskComponents = make(map[uint32]string)

// reverseWellKnownSids maps short SID names to their full string representation
var reverseWellKnownSids = make(map[string]string)

// reverseWellKnownAccessMasks maps access masks to their short names
var reverseWellKnownAccessMasks = make(map[string]uint32)

func init() {
	// Initialize the reverse mapping of wellKnownSids
	for k, v := range wellKnownSids {
		reverseWellKnownSids[v] = k
	}

	// Initialize the reverse mapping of wellKnownAccessMasks
	for k, v := range wellKnownAccessMasks {
		reverseWellKnownAccessMasks[v] = k
	}

	// Initialize the reverse mapping of accessMaskComponents
	for k, v := range accessMaskComponents {
		reversedAccessMaskComponents[v] = k
	}
}

// ace represents a Windows Access Control Entry (ACE)
// The ace structure is used in the ACL data structure to specify access control information for an object.
// It contains information such as the type of ace, the access control information, and the SID of the trustee.
// See https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-ace
type ace struct {
	// header is the ACE header, which contains the type of ACE, flags, and size.
	header *aceHeader
	// accessMask is the access mask containing the access rights that are being granted or denied.
	// It is a combination of the standard access rights and the specific rights defined by the object.
	// See https://docs.microsoft.com/en-us/windows/win32/consent/access-mask-format
	accessMask uint32
	// sid is the sid of the trustee, which is the user or group that the ACE is granting or denying access to.
	sid *sid
}

// accessString returns a string representation of the access mask, checking for well-known combinations first
func (e *ace) accessString() string {
	var accessStr string
	if value, ok := wellKnownAccessMasks[e.accessMask]; ok {
		accessStr = value
	} else {
		maskComponents, remainingMask := decomposeAccessMask(e.accessMask)
		accessStr = strings.Join(maskComponents, "")
		if remainingMask != 0 {
			accessStr = fmt.Sprintf("0x%08X", e.accessMask)
		}
	}

	return accessStr
}

// Binary converts an ACE structure to its binary representation following Windows format.
// The binary format is:
// - ACE Header:
//   - AceType (1 byte)
//   - AceFlags (1 byte)
//   - AceSize (2 bytes, little-endian)
//
// - AccessMask (4 bytes, little-endian)
// - SID in binary format (variable size)
func (e *ace) Binary() []byte {
	// Validate ACE structure
	if e == nil {
		panic("cannot convert nil ACE to binary")
	}
	if e.header == nil {
		panic("cannot convert ACE with nil header to binary")
	}
	if e.sid == nil {
		panic("cannot convert ACE with nil SID to binary")
	}

	// Convert SID to binary first to get its size
	sidBinary := e.sid.Binary()

	// Calculate total ACE size: 4 (header) + 4 (access mask) + len(sidBinary)
	aceSize := 4 + 4 + len(sidBinary)
	if aceSize > 65535 { // Check if size fits in uint16
		panic("ACE size exceeds maximum size of 65535 bytes")
	}

	// Validate that the calculated size matches the header size
	if uint16(aceSize) != e.header.aceSize {
		panic("calculated ACE size doesn't match header size")
	}

	// Create result buffer
	result := make([]byte, aceSize)

	// Set ACE header
	result[0] = e.header.aceType
	result[1] = e.header.aceFlags
	binary.LittleEndian.PutUint16(result[2:4], uint16(aceSize))

	// Set access mask (4 bytes, little-endian)
	binary.LittleEndian.PutUint32(result[4:8], e.accessMask)

	// Copy SID binary representation
	copy(result[8:], sidBinary)

	return result
}

// flagsString converts the ACE flags to string
func (e *ace) flagsString() string {
	var flagsStr string
	if e.header.aceType == systemAuditACEType {
		if e.header.aceFlags&successfulAccessACE != 0 {
			flagsStr += "SA"
		}
		if e.header.aceFlags&failedAccessACE != 0 {
			flagsStr += "FA"
		}
	}

	// Add inheritance flags
	if e.header.aceFlags&objectInheritACE != 0 {
		flagsStr += "OI"
	}
	if e.header.aceFlags&containerInheritACE != 0 {
		flagsStr += "CI"
	}
	if e.header.aceFlags&inheritOnlyACE != 0 {
		flagsStr += "IO"
	}
	if e.header.aceFlags&inheritedACE != 0 {
		flagsStr += "ID"
	}

	return flagsStr
}

// String returns a string representation of the ACE.
func (e *ace) String() string {
	return fmt.Sprintf("(%s;%s;%s;;;%s)", e.typeString(), e.flagsString(), e.accessString(), e.sid.String())
}

// StringIndent returns a string representation of the ACE with the specified indentation margin.
// The margin parameter specifies the number of spaces to prepend to the output.
func (e *ace) StringIndent(margin int) string {
	eStr := fmt.Sprintf("(%s;%s;%s;;;%s)", e.typeString(), e.flagsString(), e.accessString(), e.sid.DebugString())
	return strings.Repeat(" ", margin) + eStr
}

// typeString returns a string representation of the ACE type
func (e *ace) typeString() string {
	switch e.header.aceType {
	case accessAllowedACEType:
		return "A"
	case accessDeniedACEType:
		return "D"
	case systemAuditACEType:
		return "AU"
	default:
		return fmt.Sprintf("0x%02X", e.header.aceType)
	}
}

// aceHeader represents the Windows ACE_HEADER structure, which is the header of an Access Control Entry (ACE)
// See https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/628ebb1d-c509-4ea0-a10f-77ef97ca4586
type aceHeader struct {
	// acetype - Type of ACE (ACCESS_ALLOWED_ACE_TYPE, ACCESS_DENIED_ACE_TYPE, etc.)
	aceType byte
	// aceflags (OBJECT_INHERIT_ACE, CONTAINER_INHERIT_ACE, etc.)
	aceFlags byte
	// aceSize is the total size of the ACE in bytes
	aceSize uint16
}

// acl represents the Windows Access Control List (ACL) structure
// See https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/20233ed8-a6c6-4097-aafa-dd545ed24428
type acl struct {
	// aclRevision is the revision of the ACL format. Currently, only revision 2 is supported. See
	aclRevision byte

	// Sbz1 is reserved; must be zero
	sbzl byte

	// aclSize is the size of the ACL in bytes
	aclSize uint16

	// aceCount is the number of ACEs in the ACL
	aceCount uint16

	// sbz2 is reserved; must be zero
	sbz2 uint16

	// The following fields are not part of the original structure, but they are used in conjuntion with AclType and Control to build the string representation

	// aclType is "D" for DACL, "S" for SACL.
	//
	// This field is not part of original structure, but it is used in conjuntion with Control to build the string representation
	aclType string

	// control are the Security Descriptor control flags defined in
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/7d4dac05-9cef-4563-a058-f108abecce1d
	//
	// This field is not part of original structure, but it is used in conjuntion with AclType to build the string representation
	control uint16

	// aces is the list of Access Control Entries (ACEs)
	//
	// This field is not part of original structure, but it is used to build the string representation.
	aces []ace
}

// Binary converts an ACL structure to its binary representation following Windows format.
//
// The binary format consists of:
// - ACL Header:
//   - Revision (1 byte)
//   - Sbz1 (1 byte, reserved)
//   - AclSize (2 bytes, little-endian)
//   - AceCount (2 bytes, little-endian)
//   - Sbz2 (2 bytes, reserved)
//
// - Array of ACEs in binary format (variable size)
func (a *acl) Binary() []byte {
	// Convert all ACEs to binary first to validate them and calculate total size
	aceBinaries := make([][]byte, len(a.aces))
	totalAceSize := 0

	for i := range a.aces {
		aceBinaries[i] = a.aces[i].Binary()
		totalAceSize += len(aceBinaries[i])
	}

	// Calculate total ACL size: 8 (header) + sum of ACE sizes
	aclSize := 8 + totalAceSize
	if aclSize > 65535 { // Check if size fits in uint16
		panic(fmt.Errorf("ACL size %d exceeds maximum size of 65535 bytes", aclSize))
	}

	// Validate that calculated size matches the ACL size field
	if uint16(aclSize) != a.aclSize {
		panic(fmt.Errorf("calculated ACL size %d doesn't match header size %d", aclSize, a.aclSize))
	}

	// Validate ACE count
	if uint16(len(a.aces)) != a.aceCount {
		panic(fmt.Errorf("actual ACE count %d doesn't match header count %d", len(a.aces), a.aceCount))
	}

	// Create result buffer
	result := make([]byte, aclSize)

	// Set ACL header
	result[0] = a.aclRevision
	result[1] = a.sbzl // Reserved byte
	binary.LittleEndian.PutUint16(result[2:4], uint16(aclSize))
	binary.LittleEndian.PutUint16(result[4:6], uint16(len(a.aces)))
	binary.LittleEndian.PutUint16(result[6:8], a.sbz2) // Reserved bytes

	// Copy each ACE's binary representation
	offset := 8
	for _, aceBinary := range aceBinaries {
		copy(result[offset:], aceBinary)
		offset += len(aceBinary)
	}

	return result
}

// FlagsString returns a string representation of the ACL flags.
// It constructs the flag string based on the ACL type (DACL or SACL) and the control flags.
// The returned string format is "Type:Flags", where Type is either "D" for DACL or "S" for SACL,
// and Flags is a combination of the following:
//   - "P" for Protected
//   - "AI" for Auto-Inherited
//   - "AR" for Auto-Inherit Required
//   - "R" for Read-Only
//
// If no flags are set, it returns just the ACL type.
func (a *acl) FlagsString() string {
	var aclFlags []string
	if a.aclType == "D" {
		if a.control&seDACLProtected != 0 {
			aclFlags = append(aclFlags, "P")
		}
		if a.control&seDACLAutoInherited != 0 {
			aclFlags = append(aclFlags, "AI")
		}
		if a.control&seDACLAutoInheritRe != 0 {
			aclFlags = append(aclFlags, "AR")
		}
		if a.control&seDACLDefaulted != 0 {
			aclFlags = append(aclFlags, "R")
		}
	} else if a.aclType == "S" {
		if a.control&seSACLProtected != 0 {
			aclFlags = append(aclFlags, "P")
		}
		if a.control&seSACLAutoInherited != 0 {
			aclFlags = append(aclFlags, "AI")
		}
		if a.control&seSACLAutoInheritRe != 0 {
			aclFlags = append(aclFlags, "AR")
		}
		if a.control&seSACLDefaulted != 0 {
			aclFlags = append(aclFlags, "R")
		}
	}

	return strings.Join(aclFlags, "")
}

func (a *acl) String() string {
	result := a.FlagsString()

	var aces []string
	for _, ace := range a.aces {
		aces = append(aces, ace.String())
	}

	return result + strings.Join(aces, "")
}

// StringIndent returns a string representation of the ACL with the specified indentation margin.
// It formats the ACL flags and each ACE on separate lines, with ACEs indented 4 spaces further
// than the margin parameter.
//
// Parameters:
//   - margin: number of spaces to prepend to each line
//
// Returns a multi-line string with the ACL flags followed by indented ACEs.
func (a *acl) StringIndent(margin int) string {
	marginStr := strings.Repeat(" ", margin)
	bldr := strings.Builder{}
	bldr.WriteString(marginStr + a.FlagsString() + "\n")
	for _, ace := range a.aces {
		bldr.WriteString(ace.StringIndent(margin+4) + "\n")
	}
	return bldr.String()
}

// SecurityDescriptor represents the Windows SECURITY_DESCRIPTOR structure.
//
// A security descriptor is a data structure that contains the security
// information associated with a securable object, such as a file, registry
// key, or network share. It includes an owner SID, a primary group SID,
// a discretionary access control list (DACL) that specifies the access
// rights allowed or denied to specific users or groups, and a system
// access control list (SACL) that specifies the types of auditing that
// are to be generated for specific users or groups.
//
// See:
//   - https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/7d4dac05-9cef-4563-a058-f108abecce1d
//   - https://learn.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-control
type SecurityDescriptor struct {
	// revision of the security descriptor format.
	// Valid values are 1 (for Windows XP and later) and 2 (for Windows 2000).
	// The revision determines the offset of the owner and group SIDs:
	// in revision 1, the offset is 4 bytes, and in revision 2, the offset is 8 bytes.
	revision byte

	// sbzl is Reserved; must be zero
	sbzl byte

	// control flags
	// The control field specifies the type of security descriptor and other flags.
	control uint16

	// Offset of owner SID in bytes relative to start of security descriptor
	ownerOffset uint32

	// Offset of group SID in bytes relative to start of security descriptor
	groupOffset uint32

	// Offset of SACL in bytes relative to start of security descriptor
	saclOffset uint32

	// Offset of DACL in bytes relative to start of security descriptor
	daclOffset uint32

	// The following fields are not part of original structure but are needed for string representation

	// ownerSID is the Owner of the SID.
	//
	// This field is not part of original structure, but it is used to build the string representation.
	ownerSID *sid

	// groupSID is the Group of the SID.
	//
	// This field is not part of original structure, but it is used to build the string representation.
	groupSID *sid

	// sacl is the System Access Control List (SACL).
	//
	// The sacl is used to specify the types of auditing that are to be generated for specific users or groups.
	// It is used to generate audit logs when a user or group attempts to access a securable object in a certain way.
	//
	// This field is not part of original structure, but it is used to build the string representation.
	sacl *acl

	// dacl is the Discretionary Access Control List (DACL).
	//
	// The dacl controls access to the securable object based on the user or group that is accessing it.
	//
	// This field is not part of original structure, but it is used to build the string representation.
	dacl *acl
}

// Binary converts a SecurityDescriptor structure to its binary representation in self-relative format.
// The binary format consists of:
// - Fixed part:
//   - Revision (1 byte)
//   - Sbz1 (1 byte, reserved)
//   - Control (2 bytes, little-endian)
//   - OwnerOffset (4 bytes, little-endian)
//   - GroupOffset (4 bytes, little-endian)
//   - SaclOffset (4 bytes, little-endian)
//   - DaclOffset (4 bytes, little-endian)
//
// - Variable part (in canonical order):
//   - Owner SID
//   - Group SID
//   - SACL
//   - DACL
func (sd *SecurityDescriptor) Binary() []byte {
	// Force SE_SELF_RELATIVE flag as we're creating a self-relative security descriptor
	sd.control |= seSelfRelative

	// Convert all components to binary first to calculate total size and validate
	var ownerBinary, groupBinary, saclBinary, daclBinary []byte

	// Convert Owner SID if present
	if sd.ownerSID != nil {
		ownerBinary = sd.ownerSID.Binary()
	}

	// Convert Group SID if present
	if sd.groupSID != nil {
		groupBinary = sd.groupSID.Binary()
	}

	// Convert SACL if present and control flags indicate it should be
	if sd.sacl != nil {
		if sd.control&seSACLPresent == 0 {
			panic("SACL present but SE_SACL_PRESENT flag not set")
		}
		saclBinary = sd.sacl.Binary()
	} else if sd.control&seSACLPresent != 0 {
		panic("SE_SACL_PRESENT flag set but SACL is nil")
	}

	// Convert DACL if present and control flags indicate it should be
	if sd.dacl != nil {
		if sd.control&seDACLPresent == 0 {
			panic("DACL present but SE_DACL_PRESENT flag not set")
		}
		daclBinary = sd.dacl.Binary()
	} else if sd.control&seDACLPresent != 0 {
		panic("SE_DACL_PRESENT flag set but DACL is nil")
	}

	// Calculate total size: 20 (fixed header) + sizes of all components
	totalSize := 20 + len(ownerBinary) + len(groupBinary) + len(saclBinary) + len(daclBinary)

	// Create result buffer
	result := make([]byte, totalSize)

	// Set fixed header
	result[0] = sd.revision
	result[1] = sd.sbzl
	binary.LittleEndian.PutUint16(result[2:4], sd.control)

	// Initialize current offset for variable part
	currentOffset := 20

	// Set Owner SID and its offset if present
	if ownerBinary != nil {
		binary.LittleEndian.PutUint32(result[4:8], uint32(currentOffset))
		copy(result[currentOffset:], ownerBinary)
		currentOffset += len(ownerBinary)
	}

	// Set Group SID and its offset if present
	if groupBinary != nil {
		binary.LittleEndian.PutUint32(result[8:12], uint32(currentOffset))
		copy(result[currentOffset:], groupBinary)
		currentOffset += len(groupBinary)
	}

	// Set SACL and its offset if present
	if saclBinary != nil {
		binary.LittleEndian.PutUint32(result[12:16], uint32(currentOffset))
		copy(result[currentOffset:], saclBinary)
		currentOffset += len(saclBinary)
	}

	// Set DACL and its offset if present
	if daclBinary != nil {
		binary.LittleEndian.PutUint32(result[16:20], uint32(currentOffset))
		copy(result[currentOffset:], daclBinary)
	}

	return result
}

func (sd *SecurityDescriptor) String() string {
	var parts []string
	if sd.ownerSID != nil {
		ownerSIDString := sd.ownerSID.String()
		parts = append(parts, fmt.Sprintf("O:%s", ownerSIDString))
	}
	if sd.groupSID != nil {
		groupSIDString := sd.groupSID.String()
		parts = append(parts, fmt.Sprintf("G:%s", groupSIDString))
	}
	if sd.dacl != nil {
		daclStr := sd.dacl.String()
		parts = append(parts, fmt.Sprintf("D:%s", daclStr))
	}
	if sd.sacl != nil {
		saclStr := sd.sacl.String()
		parts = append(parts, fmt.Sprintf("S:%s", saclStr))
	}
	return strings.Join(parts, "")
}

// StringIndent returns a formatted string representation of the SecurityDescriptor with the specified
// indentation margin. It includes the control flags, owner, group, and ACLs (if present), each
// properly indented for better readability.
//
// Parameters:
//   - margin: number of spaces to prepend to each line
//
// Returns a multi-line string containing the formatted security descriptor components.
func (sd *SecurityDescriptor) StringIndent(margin int) string {
	marginStr := strings.Repeat(" ", margin)
	bldr := strings.Builder{}

	if sd.ownerSID != nil {
		bldr.WriteString(marginStr + "O: " + sd.ownerSID.String() + "\n")
	}

	if sd.groupSID != nil {
		bldr.WriteString(marginStr + "G: " + sd.groupSID.String() + "\n")
	}

	if sd.dacl != nil {
		bldr.WriteString(marginStr + "D:\n" + sd.dacl.StringIndent(margin+4) + "\n")
	}

	if sd.sacl != nil {
		bldr.WriteString(marginStr + "S:\n" + sd.sacl.StringIndent(margin+4) + "\n")
	}

	return bldr.String()
}

// sid represents a Windows Security Identifier (SID)
//
// Note: SubAuthorityCount  is needed for parsing, but once the structure is built, it can be determined from SubAuthority, hence the field is omitted in the structure
type sid struct {
	// revision indicates the revision level of the SID structure.
	// It is used to determine the format of the SID structure.
	// The current revision level is 1.
	revision byte

	// identifierAuthority is the authority part of the SID. It is a 6-byte
	// value that identifies the authority issuing the SID. The high-order
	// 2 bytes contain the revision level of the SID. The next byte is the
	// identifier authority value. The low-order 3 bytes are zero.
	identifierAuthority uint64

	// subAuthority is the sub-authority parts of the SID.
	// The number of sub-authorities is determined by SubAuthorityCount.
	// The sub-authorities are in the order they appear in the SID string
	// (i.e. S-1-5-21-a-b-c-d-e, where d and e are sub-authorities).
	// The sub-authorities are stored in little-endian order.
	// See https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-sid
	subAuthority []uint32
}

// Binary converts a SID structure to its binary representation following Windows format.
// The binary format is:
// - Revision (1 byte)
// - SubAuthorityCount (1 byte)
// - IdentifierAuthority (6 bytes, big-endian)
// - SubAuthorities (4 bytes each, little-endian)
func (s *sid) Binary() []byte {
	// Validate SID structure
	if s == nil {
		panic("cannot convert nil SID to binary")
	}

	if s.revision != 1 {
		panic(fmt.Errorf("%w: revision must be 1, was %d", ErrInvalidSIDFormat, s.revision))
	}

	// Check number of sub-authorities (maximum is 15 in Windows)
	if len(s.subAuthority) > 15 {
		panic(fmt.Errorf("%w: got %d, maximum is 15", ErrTooManySubAuthorities, len(s.subAuthority)))
	}

	// Check authority value fits in 48 bits
	if s.identifierAuthority >= 1<<48 {
		panic(fmt.Errorf("%w: value %d exceeds maximum of 2^48-1", ErrInvalidAuthority, s.identifierAuthority))
	}

	// Calculate total size:
	// 1 byte revision + 1 byte count + 6 bytes authority + (4 bytes × number of sub-authorities)
	size := 8 + (4 * len(s.subAuthority))
	result := make([]byte, size)

	// Set revision
	result[0] = s.revision

	// Set sub-authority count
	result[1] = byte(len(s.subAuthority))

	// Set authority value - convert uint64 to 6 bytes in big-endian order
	// We're using big-endian because Windows stores the authority as a 6-byte
	// value in network byte order (big-endian)
	auth := s.identifierAuthority
	for i := 7; i >= 2; i-- {
		result[i] = byte(auth & 0xFF)
		auth >>= 8
	}

	// Set sub-authorities in little-endian order
	// Windows stores these as 32-bit integers in little-endian format
	for i, subAuth := range s.subAuthority {
		offset := 8 + (4 * i)
		binary.LittleEndian.PutUint32(result[offset:], subAuth)
	}

	return result
}

// DebugString returns a string representation of the SID with additional debugging information.
// It includes the raw string representation whithout converting to well-known SID, alongside the
// final SID (in case they were different)
func (s *sid) DebugString() string {
	st := s.String()
	rs := s.rawString()

	if st != rs {
		return fmt.Sprintf("%s [%s]", st, rs)
	}

	return st
}

// Domain returns a slice of uint32 containing all sub-authorities between the first and last one.
// For example, if the SID is S-1-5-21-a-b-c-123, it will return [a,b,c].
// If there are not enough sub-authorities (less than 3), it returns an empty slice.
func (s *sid) Domain() []uint32 {
	if len(s.subAuthority) < 3 {
		return []uint32{}
	}
	return s.subAuthority[1 : len(s.subAuthority)-1]
}

func (s *sid) isGeneric() bool {
	raw := s.rawString()
	_, ok := wellKnownSids[raw]
	return ok
}

func (s *sid) rawString() string {
	authority := fmt.Sprintf("%d", s.identifierAuthority)
	if s.identifierAuthority >= 1<<32 {
		authority = fmt.Sprintf("0x%x", s.identifierAuthority)
	}

	sidStr := fmt.Sprintf("S-%d-%s", s.revision, authority)
	for _, subAuthority := range s.subAuthority {
		sidStr += fmt.Sprintf("-%d", subAuthority)
	}

	return sidStr
}

// String returns a string representation of the SID. If the SID corresponds to a well-known
// SID, the short well-known SID name will be returned instead of the full SID string.
//
// The returned string will be in the format
// "S-<revision>-<authority>-<sub-authority1>-<sub-authority2>-...-<sub-authorityN>".
// If the SID is well-known, the string will be in the format "<well-known SID name>".
func (s *sid) String() string {
	s.Validate()

	sidStr := s.rawString()

	if wk, ok := wellKnownSids[sidStr]; ok {
		return wk
	}

	// Well-known RIDs (after experimenting, only these two were converted to a short form)
	// perhaps because they belong to concrete users, while the rest represent groups
	if strings.HasPrefix(sidStr, "S-1-5-21-") && len(s.subAuthority) > 4 {
		switch s.subAuthority[len(s.subAuthority)-1] {
		case 500:
			return "LA"
		case 501:
			return "LG"
		}
	}

	return sidStr
}

func (s *sid) Validate() {
	// Check authority value fits in 48 bits
	if s.identifierAuthority >= 1<<48 {
		panic(fmt.Errorf("%w: value %d exceeds maximum of 2^48-1", ErrInvalidAuthority, s.identifierAuthority))
	}

	// Check number of sub-authorities (maximum is 15 in Windows)
	if len(s.subAuthority) > 15 {
		panic(fmt.Errorf("%w: got %d, maximum is 15", ErrTooManySubAuthorities, len(s.subAuthority)))
	}

	if s.revision != 1 {
		panic(fmt.Errorf("%w: revision must be 1, was %d", ErrInvalidSIDFormat, s.revision))
	}
}

// decomposeAccessMask breaks down an access mask into its individual components
// it also returns the mask without the components
func decomposeAccessMask(mask uint32) ([]string, uint32) {
	var components []string

	// Check components in order (least significant bits first)
	maskValues := make([]uint32, 0, len(reversedAccessMaskComponents))
	for val := range reversedAccessMaskComponents {
		maskValues = append(maskValues, val)
	}

	slices.Sort(maskValues)
	for _, val := range maskValues {
		name := reversedAccessMaskComponents[val]
		if mask&val == val {
			components = append(components, name)
			mask ^= val
		}
	}

	return components, mask
}

// composeAccessMask combines individual permission components into an access mask
// it also return the components that were unable to be combined
func composeAccessMask(components []string) (uint32, []string) {
	var remaining []string
	var mask uint32
	for _, code := range components {
		if val, ok := accessMaskComponents[code]; ok {
			mask |= val
		} else {
			remaining = append(remaining, code)
		}
	}
	return mask, remaining
}
