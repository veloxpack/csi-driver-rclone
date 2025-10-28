# sddl - Windows Security Descriptor Library and CLI Tool

A cross-platform Go library and command-line tool for working with Windows Security Descriptors, providing conversion between binary and SDDL (Security Descriptor Definition Language) string formats.

## Features

- Convert between binary and SDDL string formats
- Read security descriptors directly from files on Windows systems
- Support for all Security Descriptor components:
  - Owner and Group SIDs
  - DACLs and SACLs
  - All standard ACE types
  - Inheritance flags
  - ACL control flags
- Translation of well-known SIDs to aliases (e.g., "SY" for SYSTEM)
- Translation of common access masks to symbolic form (e.g., "FA" for Full Access)
- Cross-platform library functionality
- Windows-specific features when available
- Pure Go implementation with minimal dependencies

## CLI Tool Usage

The command-line tool provides several modes of operation:

### Basic Usage

```bash
# Convert base64-encoded binary descriptor to SDDL string (reads from stdin, writes to stdout)
echo "AQAAgBQAAAAkAAAAAAAAABAAAAAQAAQACAAEABIAAAA=" | sddl -i binary -o string > output.txt

# Convert SDDL string to base64-encoded binary (reads from stdin, writes to stdout)
echo "O:SYG:SY" | sddl -i string -o binary > binary_output.txt

# Read security descriptors from files (Windows only, filenames from stdin)
echo "C:\Windows\notepad.exe" | sddl -file > security_descriptors.txt
```

### Input/Output Formats

- `-i format`: Input format, either 'binary' (base64 encoded) or 'string' (SDDL)
- `-o format`: Output format, either 'binary' (base64 encoded) or 'string' (SDDL)
- `-file`: Process input as filenames and read their security descriptors (Windows only)
- `-debug`: Prints the result in a human-readable format (applies only when `-o string` is used)

### Examples

```bash
# Convert binary to SDDL
echo "AQAAgBQAAAAkAAAAAAAAABAAAAAQAAQACAAEABIAAAA=" | sddl -i binary -o string
# Output: O:SYG:SY

# Convert SDDL to binary
echo "O:SYG:SY" | sddl -i string -o binary
# Output: AQAAgBQAAAAkAAAAAAAAABAAAAAQAAQACAAEABIAAAA=

# Get security descriptor from files (Windows only)
echo "C:\Windows\notepad.exe" | sddl -file -o string
# Output: O:SYG:BAD:(A;;FA;;;SY)
```

### Processing Rules

- Reads input line by line from stdin
- Each line should contain either a single security descriptor or filename
- Empty lines are ignored
- Processing continues even if some lines fail
- Errors are reported to stderr with line numbers
- Results are written to stdout, one per line

## Library Usage

### Installation

```bash
go get github.com/cloudsoda/sddl
```

### Basic Usage

```go
import "github.com/cloudsoda/sddl"

// Parse binary security descriptor
sd, err := sddl.FromBinary(binaryData)
if err != nil {
    // Handle error
}
sddlString, err := sd.String()

// Parse SDDL string
sd, err := sddl.FromString("O:SYG:BAD:(A;;FA;;;SY)")
if err != nil {
    // Handle error
}
binaryData, err := sd.Binary()
```

### Windows-Specific Features

```go
// Windows only: Get security descriptor from file
sddlString, err := GetFileSDString("C:\\Windows\\notepad.exe")

// Windows only: Get binary security descriptor from file
base64Data, err := GetFileSecurityBase64("C:\\Windows\\notepad.exe")
```

## SDDL Format

Security descriptors in SDDL format follow this structure:
```
O:owner_sidG:group_sidD:dacl_flagsS:sacl_flags
```

Components:
- Owner SID (`O:`): Specifies the owner
- Group SID (`G:`): Specifies the primary group
- DACL (`D:`): Discretionary Access Control List
- SACL (`S:`): System Access Control List

### ACL Format

ACLs contain flags and a list of ACEs (Access Control Entries):
```
D:flags(ace1)(ace2)...(aceN)
```

ACL Flags:
- `P`: Protected
- `AI`: Auto-inherited
- `AR`: Auto-inherit required
- `NO`: No propagate inherit

### ACE Format

Each ACE follows this format:
```
(ace_type;ace_flags;rights;object_guid;inherit_object_guid;account_sid)
```

Components:
- `ace_type`: Type of ACE (e.g., "A" for Allow, "D" for Deny)
- `ace_flags`: Inheritance flags (e.g., "CI" for Container Inherit)
- `rights`: Access rights (e.g., "FA" for Full Access)
- `account_sid`: Security identifier for the trustee

## Error Handling

The library provides detailed error information for various scenarios:

- Invalid security descriptor structure
- Malformed SIDs
- Invalid ACL or ACE formats
- Base64 decoding errors
- File access errors (Windows-specific features)

Errors include context about where in the parsing process they occurred to aid in debugging.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
